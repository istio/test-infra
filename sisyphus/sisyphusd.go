// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sisyphus

import (
	"context"
	"fmt"
	"log"
	"time"

	"istio.io/test-infra/toolbox/util"
)

const (
	// DefaultNumRerun is the default number of reruns for each failured jobs
	DefaultNumRerun = 2
	// DefaultCatchFlakesByRun defines if reruns are triggered by default
	DefaultCatchFlakesByRun = true
)

var (
	pendingJobTimeout = 120 * time.Minute
	// DefaultPollGapDuration is the default time that sisyphus waits between
	// two checks on jobs
	DefaultPollGapDuration = 300 * time.Second
)

type jobStatus struct {
	// Assume job name unique
	name string
	// Monotonically increasing, the run number identifies a particular run of the job
	// lastCheckedRunNo is the latest run number we have checked in most recent poll
	// Used to ensure we check each run exactly once
	lastCheckedRunNo int
	// Jobs may finish in different order in which they start
	// Key: run number whose result still pending
	// Value: time when the first attempt to fetch result was made
	pendingFirstRuns map[int]time.Time
	// Key: Commit SHA from which reruns are triggered
	// Value: FlakeStat
	rerunJobStats map[string]*FlakeStat
}

type failure struct {
	jobName string
	runNo   int
}

// FlakeStat records the stats from flakiness detection by multiple reruns
type FlakeStat struct {
	TestName   string `json:"testName"`
	SHA        string `json:"sha"`
	TotalRerun int    `json:"totalRerun"`
	Failures   int    `json:"failures"`
}

// Daemon defines the long-running service that sisyphus provides
type Daemon struct {
	jobsWatched      []*jobStatus
	ci               CI
	storage          Storage
	pollGapDuration  time.Duration
	numRerun         int
	catchFlakesByRun bool
	// optional
	alert           *Alert
	branchProtector *branchProtector
}

// Config is the optional configuration to daemon
type Config struct {
	PollGapDuration  time.Duration
	NumRerun         int
	CatchFlakesByRun bool
}

func newDaemon(protectedJobs []string, cfg *Config, storage Storage) *Daemon {
	var jobsWatched []*jobStatus
	for _, jobName := range protectedJobs {
		jobsWatched = append(jobsWatched, &jobStatus{
			name:             jobName,
			lastCheckedRunNo: 0,
			pendingFirstRuns: make(map[int]time.Time),
			rerunJobStats:    make(map[string]*FlakeStat),
		})
	}
	daemon := Daemon{
		jobsWatched:      jobsWatched,
		storage:          storage,
		pollGapDuration:  DefaultPollGapDuration,
		numRerun:         DefaultNumRerun,
		catchFlakesByRun: DefaultCatchFlakesByRun,
	}
	if cfg != nil {
		if cfg.PollGapDuration.Nanoseconds() > 0 {
			daemon.pollGapDuration = cfg.PollGapDuration
		}
		if cfg.NumRerun > 0 {
			daemon.numRerun = cfg.NumRerun
		}
		daemon.catchFlakesByRun = cfg.CatchFlakesByRun
	}
	return &daemon
}

// NewDaemonUsingProw creates a Daemon that uses Prow as the the CI system.
// It signature ensures proper setup of a Prow client.
func NewDaemonUsingProw(
	protectedJobs, presubmitJobs []string,
	prowProject, prowZone, gubernatorURL, gcsBucket string,
	client util.IGCSClient,
	storage Storage,
	cfg *Config) *Daemon {
	daemon := newDaemon(protectedJobs, cfg, storage)
	prowAccessor := NewProwAccessor(prowProject, prowZone, gubernatorURL, gcsBucket, client)
	prowAccessor.RegisterPresubmitJobs(presubmitJobs)
	daemon.ci = prowAccessor
	return daemon
}

// GetConfig returns the Config in use by d
func (d *Daemon) GetConfig() *Config {
	return &Config{
		PollGapDuration:  d.pollGapDuration,
		NumRerun:         d.numRerun,
		CatchFlakesByRun: d.catchFlakesByRun,
	}
}

// SetAlert activates email alerts to receiverAddr when jobs failed
func (d *Daemon) SetAlert(gmailAppPass, identity, senderAddr, receiverAddr string,
	alertConfig *AlertConfig) error {
	var err error
	d.alert, err = NewAlert(gmailAppPass, identity, senderAddr, receiverAddr, alertConfig)
	return err
}

// SetProtectedBranch disables auto merging PRs on protected branch if jobs failed
func (d *Daemon) SetProtectedBranch(owner, token, repo, branch string) {
	d.branchProtector = newBranchProtector(owner, token, repo, branch)
}

// Start activates the daemon to start polling Prow results
func (d *Daemon) Start(ctx context.Context) {
	for {
		d.iterate()
		log.Printf("Sleeping for %d", d.pollGapDuration)
		select {
		case <-ctx.Done():
			log.Printf("Received cancel signal, exiting")
			return
		case <-time.After(d.pollGapDuration):
			fmt.Println("------ Resume ------")
		}
	}
}

func (d *Daemon) iterate() {
	newFailures := d.checkOnJobsWatched()
	if d.alert != nil {
		if newFailures != nil {
			log.Printf("%d tests failed in last circle", len(newFailures))
			if err := d.alert.Send(d.formatMessage(newFailures)); err != nil {
				log.Printf("Unable to send alerts: %v", err)
			}
		} else {
			log.Printf("No new tests failed in last circle.")
		}
	}
	if d.branchProtector != nil {
		d.branchProtector.process(newFailures)
	}
}

// From last check onward, check all the runs of jobsWatched.
// Record failure if it hasn't been recorded and update latestRun if necessary
// Returns an array of postSubmitJob since the same job might fail at multiple runs
func (d *Daemon) checkOnJobsWatched() []failure {
	var newFailures []failure
	for _, job := range d.jobsWatched {
		CurrentRunNo, err := d.ci.GetLatestRun(job.name)
		if err != nil {
			log.Printf("Failed to get last run number of %s: %v\n", job.name, err)
			continue
		}
		log.Printf("Job: [%s] \t Current Run No: [%d] \t Previously Checked: [%d]\n",
			job.name, CurrentRunNo, job.lastCheckedRunNo)
		// Avoid pulling entire history when the daemon has just started
		if job.lastCheckedRunNo == 0 {
			job.lastCheckedRunNo = CurrentRunNo - 1
		}
		// Clone a slice of keys in map to avoid editing while iterating
		// Then check previously pending runs to see if results are ready
		pendingRuns := make([]int, 0, len(job.pendingFirstRuns))
		for runNo := range job.pendingFirstRuns {
			pendingRuns = append(pendingRuns, runNo)
		}
		log.Printf("Checking previously pending run numbers: %v\n", pendingRuns)
		for _, runNo := range pendingRuns {
			if f := d.fetchAndProcessResult(job, runNo); f != nil {
				newFailures = append(newFailures, *f)
			}
		}
		// Check new runs since last check
		log.Printf("Checking new run numbers since last check\n")
		for runNo := job.lastCheckedRunNo + 1; runNo <= CurrentRunNo; runNo++ {
			if f := d.fetchAndProcessResult(job, runNo); f != nil {
				newFailures = append(newFailures, *f)
			}
		}
		log.Printf("Finished checking [%s]\n", job.name)
		job.lastCheckedRunNo = CurrentRunNo
	}
	return newFailures
}

func (d *Daemon) formatMessage(failures []failure) (mess string) {
	for _, f := range failures {
		mess += fmt.Sprintf("%s failed: %s\n\n",
			f.jobName, d.ci.GetDetailsURL(f.jobName, f.runNo))
	}
	return
}

func (d *Daemon) fetchAndProcessResult(job *jobStatus, runNo int) *failure {
	result, err := d.ci.GetResult(job.name, runNo)
	if err != nil {
		log.Printf("Prow result still pending for %s/%d\n", job.name, runNo)
		if firstTryTime, exists := job.pendingFirstRuns[runNo]; exists {
			if time.Since(firstTryTime).Nanoseconds() > pendingJobTimeout.Nanoseconds() {
				log.Printf("Give up polling %s/%d\n", job.name, runNo)
				delete(job.pendingFirstRuns, runNo)
			}
		} else {
			job.pendingFirstRuns[runNo] = time.Now()
		}
		return nil
	}
	if _, exists := job.pendingFirstRuns[runNo]; exists {
		log.Printf("Former pending prow result is available for %s/%d\n", job.name, runNo)
		delete(job.pendingFirstRuns, runNo)
	}
	log.Printf("%s/%d -- Passed? [%t]\n", job.name, runNo, result.Passed)
	return d.processResult(job, runNo, result)
}

func (d *Daemon) processResult(job *jobStatus, runNo int, result *Result) *failure {
	if d.catchFlakesByRun {
		if flakeStatPtr, exists := job.rerunJobStats[result.SHA]; exists {
			flakeStatPtr.TotalRerun++
			if result.Passed {
				log.Printf("Job [%s] at sha [%s] is now green\n", job.name, result.SHA)
				if err := d.storage.Store(job.name, result.SHA, *flakeStatPtr); err != nil {
					log.Printf("Failed to store flakeStat: %v\n", err)
				}
				// delete result.SHA from job.rerunJobStats since all reruns have finished
				delete(job.rerunJobStats, result.SHA)
			} else {
				flakeStatPtr.Failures++
				if flakeStatPtr.TotalRerun == d.numRerun {
					log.Printf("All %d reruns on job [%s] at sha [%s] have failed\n",
						d.numRerun, job.name, result.SHA)
				} else {
					// flakeStatPtr.TotalRerun < d.numRerun
					// start the next rerun
					log.Printf("Starting the %d-th rerun on job [%s] at sha [%s]",
						flakeStatPtr.TotalRerun+1, job.name, result.SHA)
					if err := d.ci.Rerun(job.name, runNo); err != nil {
						log.Printf("failed when starting reruns on [%s]: %v\n", job.name, err)
					}
				}
			}
		} else if !result.Passed { // no reruns exist on this SHA
			log.Printf("Starting new rerun task on job [%s] at sha [%s]\n", job.name, result.SHA)
			job.rerunJobStats[result.SHA] = &FlakeStat{
				TestName: job.name,
				SHA:      result.SHA,
			}
			// start the first rerun
			log.Printf("Starting the first rerun on job [%s] at sha [%s]", job.name, result.SHA)
			if err := d.ci.Rerun(job.name, runNo); err != nil {
				log.Printf("failed when starting reruns on [%s]: %v\n", job.name, err)
			}
		}
	}
	if !result.Passed {
		return &failure{
			jobName: job.name,
			runNo:   runNo,
		}
	}
	return nil
}

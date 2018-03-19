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
	"fmt"
	"log"
	"time"
)

const (
	DefaultNumRerun         = 3
	DefaultCatchFlakesByRun = true
)

var (
	pendingJobTimeout      = 60 * time.Minute
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
	TestName           string `json:"testName"`
	SHA                string `json:"sha"`
	TotalRerun         int    `json:"totalRerun"`
	Failures           int    `json:"failures"`
	ParentJobTimeStamp uint32 `json:"parentJobTimeStamp"`
}

type sisyphusDaemon struct {
	jobsWatched      []*jobStatus
	prowAccessor     IProwAccessor
	storage          ISisyphusStorage
	pollGapDuration  time.Duration
	numRerun         int
	catchFlakesByRun bool
	// optional
	alert           *alert
	branchProtector *branchProtector
	exitSignal      chan bool
}

// SisyphusConfig is the optional configuration to SisyphusDaemon
type SisyphusConfig struct {
	PollGapDuration  time.Duration
	NumRerun         int
	CatchFlakesByRun bool
}

// SisyphusDaemon creates a sisyphusDaemon
func SisyphusDaemon(protectedJobs []string,
	prowProject, prowZone, gubernatorURL, gcsBucket string, cfg *SisyphusConfig) *sisyphusDaemon {
	var jobsWatched []*jobStatus
	for _, jobName := range protectedJobs {
		jobsWatched = append(jobsWatched, &jobStatus{
			name:             jobName,
			lastCheckedRunNo: 0,
			pendingFirstRuns: make(map[int]time.Time),
			rerunJobStats:    make(map[string]*FlakeStat),
		})
	}
	daemon := sisyphusDaemon{
		jobsWatched:      jobsWatched,
		prowAccessor:     NewProwAccessor(prowProject, prowZone, gubernatorURL, gcsBucket),
		storage:          NewSisyphusStorage(),
		pollGapDuration:  DefaultPollGapDuration,
		numRerun:         DefaultNumRerun,
		catchFlakesByRun: DefaultCatchFlakesByRun,
		exitSignal:       make(chan bool, 1), // default channel never receives, hence not blocking
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

// returns the SisyphusConfig in use by d
func (d *sisyphusDaemon) GetSisyphusConfig() *SisyphusConfig {
	return &SisyphusConfig{
		PollGapDuration:  d.pollGapDuration,
		NumRerun:         d.numRerun,
		CatchFlakesByRun: d.catchFlakesByRun,
	}
}

// SetAlert activates email alerts to receiverAddr when jobs failed
func (d *sisyphusDaemon) SetAlert(gmailAppPass, identity, senderAddr, receiverAddr string,
	alertConfig *AlertConfig) error {
	var err error
	d.alert, err = NewAlert(gmailAppPass, identity, senderAddr, receiverAddr, alertConfig)
	return err
}

// SetProtectedBranch disables auto merging PRs on protected branch if jobs failed
func (d *sisyphusDaemon) SetProtectedBranch(owner, token, repo, branch string) {
	d.branchProtector = newBranchProtector(owner, token, repo, branch)
}

// SetExitSignal sets a channel on sisyphusDaemon
func (d *sisyphusDaemon) SetExitSignal(channel chan bool) {
	d.exitSignal = channel
}

// Start activates the SisyphusDaemon to start polling Prow results
func (d *sisyphusDaemon) Start() {
	for {
		newFailures := d.checkOnJobsWatched()
		if d.alert != nil {
			if newFailures != nil {
				log.Printf("%d tests failed in last circle", len(newFailures))
				d.alert.Send(d.formatMessage(newFailures))
			} else {
				log.Printf("No new tests failed in last circle.")
			}
		}
		if d.branchProtector != nil {
			d.branchProtector.process(newFailures)
		}
		log.Printf("Sleeping for %d", d.pollGapDuration)
		select {
		case <-d.exitSignal:
			log.Printf("Received from exitSignal channel, exiting")
			return
		case <-time.After(d.pollGapDuration):
			fmt.Println("------ Resume ------")
		}
	}
}

// From last check onward, check all the runs of jobsWatched.
// Record failure if it hasn't been recorded and update latestRun if necessary
// Returns an array of postSubmitJob since the same job might fail at multiple runs
func (d *sisyphusDaemon) checkOnJobsWatched() []failure {
	var newFailures []failure
	for _, job := range d.jobsWatched {
		CurrentRunNo, err := d.prowAccessor.GetLatestRun(job.name)
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
			if f := d.fetchAndProcessProwResult(job, runNo); f != nil {
				newFailures = append(newFailures, *f)
			}
		}
		// Check new runs since last check
		log.Printf("Checking new run numbers since last check\n")
		for runNo := job.lastCheckedRunNo + 1; runNo <= CurrentRunNo; runNo++ {
			if f := d.fetchAndProcessProwResult(job, runNo); f != nil {
				newFailures = append(newFailures, *f)
			}
		}
		log.Printf("Finished checking [%s]\n", job.name)
		job.lastCheckedRunNo = CurrentRunNo
	}
	return newFailures
}

func (d *sisyphusDaemon) formatMessage(failures []failure) (mess string) {
	for _, f := range failures {
		mess += fmt.Sprintf("%s failed: %s/%s/%d\n\n",
			f.jobName, d.prowAccessor.GetGubernatorURL(), f.jobName, f.runNo)
	}
	return
}

func (d *sisyphusDaemon) fetchAndProcessProwResult(job *jobStatus, runNo int) *failure {
	prowResult, err := d.prowAccessor.GetProwResult(job.name, runNo)
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
	log.Printf("%s/%d -- Passed? [%t]\n", job.name, runNo, prowResult.Passed)
	return d.processProwResult(job, runNo, prowResult)
}

func (d *sisyphusDaemon) processProwResult(job *jobStatus, runNo int, prowResult *ProwResult) *failure {
	if d.catchFlakesByRun {
		resultSHA := prowResult.Metadata.RepoCommit
		if flakeStatPtr, exists := job.rerunJobStats[resultSHA]; exists {
			flakeStatPtr.TotalRerun++
			if !prowResult.Passed {
				flakeStatPtr.Failures++
			}
			if flakeStatPtr.TotalRerun == d.numRerun {
				log.Printf("All reruns on job [%s] at sha [%s] have finished\n", job.name, resultSHA)
				if err := d.storage.Store(job.name, resultSHA, *flakeStatPtr); err != nil {
					log.Printf("Failed to store flakeStat: %v\n", err)
				}
				// delete resultSHA from job.rerunJobStats since all reruns have finished
				delete(job.rerunJobStats, resultSHA)
			}
		} else { // no reruns exist on this SHA
			if !prowResult.Passed {
				log.Printf("Starting new rerun task on job [%s] at sha [%s]\n", job.name, resultSHA)
				job.rerunJobStats[resultSHA] = &FlakeStat{
					TestName:           job.name,
					SHA:                resultSHA,
					ParentJobTimeStamp: prowResult.TimeStamp,
				}
				if err := d.prowAccessor.Rerun(job.name, runNo, d.numRerun); err != nil {
					log.Printf("failed when starting reruns on [%s]: %v\n", job.name, err)
				}
			}
		}
	}
	if !prowResult.Passed {
		return &failure{
			jobName: job.name,
			runNo:   runNo,
		}
	}
	return nil
}

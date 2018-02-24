// Copyright 2017 Istio Authors
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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"path/filepath"
	"strconv"
	"time"

	u "istio.io/test-infra/toolbox/util"
)

// TODO reorder functions for better readability
// TODO include in email about the flakeStats
// TODO generate JUnit XML and get testcase level results

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

var (
	gcsClient  *u.GCSClient
	githubClnt *u.GithubClient

	jobsWatched       []*jobStatus
	protectedJobs     = []string{"istio-postsubmit", "e2e-suite-rbac-auth", "e2e-suite-rbac-no_auth"}
	receivers         = []string{oncallMaillist}
	gmailAppPass      string
	location          *time.Location
	pendingJobTimeout = 60 * time.Minute
)

func init() {
	flag.Parse()
	gcsClient = u.NewGCSClient()
	var err error
	if *guardProtectedBranch {
		token, err := u.GetAPITokenFromFile(*tokenFile)
		if err != nil {
			log.Fatalf("Error accessing user supplied token_file: %v\n", err)
		}
		githubClnt = u.NewGithubClient(*owner, token)
	}
	// TODO set to true
	if *emailSending {
		if gmailAppPass, err = u.GetPasswordFromFile(*gmailAppPassFile); err != nil {
			log.Fatalf("Error accessing gmail app password: %v", err)
		}
	}
	// Ensure all reruns commands are issued to Prow
	if _, err := u.Shell(`gcloud container clusters get-credentials prow \
		--project=%s --zone=%s`, *gcpProject, *prowZone); err != nil {
		log.Fatalf("Unable to switch to prow cluster: %v\n", err)
	}
	for _, jobName := range protectedJobs {
		jobsWatched = append(jobsWatched, &jobStatus{
			name:             jobName,
			lastCheckedRunNo: 0,
			pendingFirstRuns: make(map[int]time.Time),
			rerunJobStats:    make(map[string]*FlakeStat),
		})
	}
	if location, err = time.LoadLocation(losAngeles); err != nil {
		log.Fatalf("Error loading time location")
	}
}

func main() {
	for {
		newFailures := checkOnJobsWatched()
		if *emailSending {
			if len(newFailures) > 0 {
				log.Printf("%d tests failed in last circle", len(newFailures))
				sendMessage(formatMessage(newFailures))
			} else {
				log.Printf("No new tests failed in last circle.")
			}
		}
		if *guardProtectedBranch {
			if len(newFailures) > 0 {
				u.BlockMergingOnBranch(githubClnt, *protectedRepo, *protectedBranch)
			} else {
				u.UnBlockMergingOnBranch(githubClnt, *protectedRepo, *protectedBranch)
			}
		}
		log.Printf("Sleeping for %d seconds", *interval)
		time.Sleep(time.Duration(*interval) * time.Second)
	}
}

func rerun(job *jobStatus, runNo int) error {
	cfg, err := getProwJobConfig(job, runNo)
	if err != nil {
		return err
	}
	if err = triggerConcurrentReruns(job, cfg); err != nil {
		return err
	}
	return nil
}

// TODO (chx) maxConcurrentJobs control and make it a util
// Only log this error and continue. Should never exit process.
func triggerConcurrentReruns(job *jobStatus, cfg ProwJobConfig) error {
	log.Printf("Rerunning %s\n", job.name)
	recess := 1 * time.Minute
	maxRetry := 3
	for i := 0; i < *numRerun; i++ {
		if err := u.Retry(recess, maxRetry, func() error {
			_, e := u.Shell(
				"kubectl create -f \"https://prow.istio.io/rerun?prowjob=%s\"", cfg.Node)
			return e
		}); err != nil {
			log.Printf("Unable to trigger the %d-th rerun of job %v", i, job.name)
		}
	}
	return nil
}

func getProwJobConfig(job *jobStatus, runNo int) (ProwJobConfig, error) {
	cfg := ProwJobConfig{}
	jobStartedFile := filepath.Join(job.name, strconv.Itoa(runNo), startedJSON)
	StartedFileString, err := gcsClient.Read(*gcsBucket, jobStartedFile)
	if err != nil {
		return cfg, err
	}
	if err = json.Unmarshal([]byte(StartedFileString), &cfg); err != nil {
		log.Printf("Failed to unmarshal started prow job %s, %v\n", StartedFileString, err)
		return cfg, err
	}
	return cfg, nil
}

func formatMessage(failures []failure) (mess string) {
	for _, f := range failures {
		mess += fmt.Sprintf("%s failed: %s/%s/%d\n\n", f.jobName, gubernatorURL, f.jobName, f.runNo)
	}
	return
}

// Use gmail smtp server to send out email.
func sendMessage(body string) {
	msg := fmt.Sprintf("From: %s\n", sender) +
		fmt.Sprintf("To: %s\n", receivers) +
		fmt.Sprintf("Subject: %s [%s]\n\n", messageSubject,
			time.Now().In(location).Format("2006-01-02 15:04:05 PST")) +
		messagePrologue + body + messageEnding

	gmailSMTPAddr := fmt.Sprintf("%s:%d", gmailSMTPSERVER, gmailSMTPPORT)
	err := smtp.SendMail(gmailSMTPAddr, smtp.PlainAuth("istio-bot", sender, gmailAppPass, gmailSMTPSERVER),
		sender, receivers, []byte(msg))
	if err != nil {
		log.Printf("smtp error: %s\n", err)
		return
	}
	log.Printf("Alert message sent!\n")
}

// Read latest run from "latest-build.txt" of a job under gcs "istio-prow" bucket
// Example: istio-prow/istio-postsubmit/latest-build.txt
func getLatestRun(jobName string) (int, error) {
	lastBuildFile := filepath.Join(jobName, lastBuildTXT)
	latestBuildString, err := gcsClient.Read(*gcsBucket, lastBuildFile)
	if err != nil {
		return 0, err
	}
	latestBuildInt, err := strconv.Atoi(latestBuildString)
	if err != nil {
		log.Printf("Failed to convert %s to int: %v\n", latestBuildString, err)
		return 0, err
	}
	return latestBuildInt, nil
}

// From last check onward, check all the runs of protected Post-Submit tests.
// Record failure if it hasn't been recorded and update latestRun if necessary
// Returns an array of postSubmitJob since the same job might fail at multiple runs
func checkOnJobsWatched() []failure {
	newFailures := []failure{}
	for _, job := range jobsWatched {
		CurrentRunNo, err := getLatestRun(job.name)
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
			if f := fetchAndProcessProwResult(job, runNo); f != nil {
				newFailures = append(newFailures, *f)
			}
		}
		// Check new runs since last check
		log.Printf("Checking new run numbers since last check\n")
		for runNo := job.lastCheckedRunNo + 1; runNo <= CurrentRunNo; runNo++ {
			if f := fetchAndProcessProwResult(job, runNo); f != nil {
				newFailures = append(newFailures, *f)
			}
		}
		log.Printf("Finished checking [%s]\n", job.name)
		job.lastCheckedRunNo = CurrentRunNo
	}
	return newFailures
}

func fetchAndProcessProwResult(job *jobStatus, runNo int) *failure {
	prowResult, err := fetchProwResult(job.name, runNo)
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
	return processProwResult(job, runNo, prowResult)
}

func fetchProwResult(jobName string, runNo int) (*ProwResult, error) {
	target := filepath.Join(jobName, strconv.Itoa(runNo))
	jobFinishedFile := filepath.Join(target, finishedJSON)
	prowResultString, err := gcsClient.Read(*gcsBucket, jobFinishedFile)
	if err != nil {
		log.Printf("Failed to get prow job result %s: %v", target, err)
		return nil, err
	}
	prowResult := ProwResult{}
	if err = json.Unmarshal([]byte(prowResultString), &prowResult); err != nil {
		log.Printf("Failed to unmarshal prow result %s, %v", prowResultString, err)
		return nil, err
	}
	return &prowResult, nil
}

func processProwResult(job *jobStatus, runNo int, prowResult *ProwResult) *failure {
	if *catchFlakesByRun {
		resultSHA := prowResult.Metadata.RepoCommit
		if flakeStatPtr, exists := job.rerunJobStats[resultSHA]; exists {
			flakeStatPtr.TotalRerun++
			if !prowResult.Passed {
				flakeStatPtr.Failures++
			}
			if flakeStatPtr.TotalRerun == *numRerun {
				log.Printf("All reruns on job [%s] at sha [%s] have finished\n", job.name, resultSHA)
				if err := recordFlakeStatToGCS(job, *flakeStatPtr); err != nil {
					log.Printf("Failed to write flakeStat to GCS: %v\n", err)
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
				if err := rerun(job, runNo); err != nil {
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

func recordFlakeStatToGCS(job *jobStatus, newFlakeStat FlakeStat) error {
	flakeStatsFile := filepath.Join(job.name, flakeStatsJSON)
	flakeStatsString, err := gcsClient.Read(*gcsBucket, flakeStatsFile)
	if err != nil {
		return err
	}
	flakeStats, err := DeserializeFlakeStats(flakeStatsString)
	if err != nil {
		return err
	}
	flakeStats = append(flakeStats, newFlakeStat)
	updatedflakeStatsString, err := SerializeFlakeStats(flakeStats)
	if err != nil {
		return err
	}
	newflakeStatString, err := SerializeFlakeStat(newFlakeStat)
	if err != nil {
		return err
	}
	log.Printf("Writing to GCS newFlakeStat = %s\n", newflakeStatString)
	return gcsClient.Write(*gcsBucket, flakeStatsFile, updatedflakeStatsString)
}

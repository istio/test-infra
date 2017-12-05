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

	"github.com/google/go-github/github"

	u "istio.io/test-infra/toolbox/util"
)

// ProwResult matches the structure published in finished.json
type ProwResult struct {
	TimeStamp  uint32       `json:"timestamp"`
	Version    string       `json:"version"`
	Result     string       `json:"result"`
	Passed     bool         `json:"passed"`
	JobVersion string       `json:"job-version"`
	Metadata   ProwMetadata `json:"metadata"`
}

// ProwMetadata matches the structure published in finished.json
type ProwMetadata struct {
	Repo       string                 `json:"repo"`
	Repos      map[string]interface{} `json:"repos"`
	RepoCommit string                 `json:"repo-commit"`
}

type postSubmitJob struct {
	name          string
	latestRun     int
	latestRunPass bool
}

const (
	// TODO: Read from config file
	// Message Info
	sender          = "istio.testing@gmail.com"
	oncallMaillist  = "istio-oncall@googlegroups.com"
	messageSubject  = "ATTENTION - Istio Post-Submit Test Failed"
	messagePrologue = "Hi istio-oncall,\n\n" +
		"Post-Submit is failing in istio/istio, please take a look at following failure(s) and fix ASAP\n\n"
	messageEnding = "\nIf you have any questions about this message or notice inaccuracy, please contact istio-engprod@google.com."
	LosAngeles    = "America/Los_Angeles"
	// Gmail setting
	gmailSMTPSERVER = "smtp.gmail.com"
	gmailSMTPPORT   = 587

	// Prow result GCS
	lastBuildTXT  = "latest-build.txt"
	finishedJSON  = "finished.json"
	gubernatorURL = "https://k8s-gubernator.appspot.com/build/istio-prow"

	doNotMergeLabel = "PostSubmit Failed/Contact Oncall"

	// Token and password file
	tokenFileDocker        = "/etc/github/git-token"
	gmailAppPassFileDocker = "/etc/gmail/gmail-app-pass"
)

var (
	gcsClient  *u.GCSClient
	githubClnt *u.GithubClient

	gcsBucket        = flag.String("bucket", "istio-prow", "Prow artifact GCS bucket name.")
	interval         = flag.Int("interval", 5, "Check and report interval(minute)")
	debug            = flag.Bool("debug", false, "Optional to log debug message")
	owner            = flag.String("owner", "istio", "Github owner or org")
	tokenFile        = flag.String("github_token", tokenFileDocker, "Path to github token")
	gmailAppPassFile = flag.String("gmail__app_password", gmailAppPassFileDocker, "Path to gmail application password")
	protectedRepo    = flag.String("protected_repo", "istio", "Protected repo")
	protectedBranch  = flag.String("protected_branch", "master", "Protected branch")

	postSubmitJobs []*postSubmitJob

	protectedPostsubmits = []string{"istio-postsubmit", "e2e-suite-rbac-auth", "e2e-suite-rbac-no_auth"}
	receivers            = []string{oncallMaillist}
	gmailAppPass         string
	location             *time.Location
)

func init() {
	flag.Parse()

	var err error
	gcsClient = u.NewGCSClient()

	/*
		token, err := u.GetAPITokenFromFile(*tokenFile)
		if err != nil {
			log.Fatalf("Error accessing user supplied token_file: %v\n", err)
		}
	*/
	token := "b1ca9d561bb4b0f278729788ce6be21d08307ffa"
	githubClnt = u.NewGithubClient(*owner, token)

	if gmailAppPass, err = u.GetPasswordFromFile(*gmailAppPassFile); err != nil {
		log.Fatalf("Error accessing gmail app password: %v", err)
	}

	for _, t := range protectedPostsubmits {
		postSubmitJobs = append(postSubmitJobs,
			&postSubmitJob{
				name: t,
				// init to be true to avoid false negative
				latestRunPass: true,
			})
	}

	location, err = time.LoadLocation(LosAngeles)
	if err != nil {
		log.Fatalf("Error loading time location")
	}
}

func main() {
	for {
		newFailures, failedCount := updatePostSubmitStatus()
		if len(newFailures) > 0 {
			log.Printf("%d tests failed in last circle", len(newFailures))
			sendMessage(formatMessage(newFailures))
		} else {
			log.Printf("No new tests failed in last circle.")
		}

		// If any test is in failed status, trying to block, else trying to unblock
		if failedCount > 0 {
			blockPRs()
		} else {
			unBlockPRs()
		}

		log.Printf("Sleeping for %d minutes", *interval)
		time.Sleep(time.Duration(*interval) * time.Minute)
	}
}

func formatMessage(failures map[*postSubmitJob]bool) (mess string) {
	for job := range failures {
		mess += fmt.Sprintf("%s failed: %s/%s/%d\n\n", job.name, gubernatorURL, job.name, job.latestRun)
	}
	return
}

// Use gmail smtp server to send out email.
func sendMessage(body string) {
	msg := fmt.Sprintf("From: %s\n", sender) +
		fmt.Sprintf("To: %s\n", receivers) +
		fmt.Sprintf("Subject: %s [%s]\n\n", messageSubject, time.Now().In(location).Format("2006-01-02 15:04:05 PST")) +
		messagePrologue + body + messageEnding

	gmailSMTPAddr := fmt.Sprintf("%s:%d", gmailSMTPSERVER, gmailSMTPPORT)
	err := smtp.SendMail(gmailSMTPAddr, smtp.PlainAuth("istio-bot", sender, gmailAppPass, gmailSMTPSERVER),
		sender, receivers, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Print("Alert message sent!")
}

// Read latest run from "latest-build.txt" of a job under gcs "istio-prow" bucket
// Example: istio-prow/istio-postsubmit/latest-build.txt
func getLatestRun(job string) (int, error) {
	lastBuildFile := filepath.Join(job, lastBuildTXT)
	latestBuildString, err := gcsClient.GetFileFromGCSString(*gcsBucket, lastBuildFile)
	if err != nil {
		return 0, err
	}
	latestBuildInt, err := strconv.Atoi(latestBuildString)
	if err != nil {
		log.Printf("Failed to convert %s to int: %v", latestBuildString, err)
		return 0, err
	}
	return latestBuildInt, nil
}

// Check the latest run of each protected Post-Submit tests.
// Record failure if it hasn't been recorded and update latestRun if necessary
func updatePostSubmitStatus() (map[*postSubmitJob]bool, int) {
	// If a "new" failure appears in this circle, put it in and with value true
	// So if no new failure, this map should be empty --> no new notification
	newFailures := make(map[*postSubmitJob]bool)

	// Total count of jobs which is in failing status right now.
	// Including jobs failed in circles before.
	// If 0 --> unblock auto-merge
	failedCount := 0

	for _, job := range postSubmitJobs {
		latestRunNo, err := getLatestRun(job.name)
		if err != nil {
			log.Printf("Failed to get last run number of %s: %v", job.name, err)
			continue
		}

		if *debug {
			log.Printf("Job: %s -- Latest Run No: %d -- Recorded Run No: %d", job, latestRunNo, job.latestRun)
		}
		// New test finished for this job in this circle
		if latestRunNo > job.latestRun {
			prowResult, err := getProwResult(filepath.Join(job.name, strconv.Itoa(latestRunNo)))
			if err != nil {
				log.Printf("Failed to get prowResult %s/%d", job, latestRunNo)
				continue
			}
			job.latestRun = latestRunNo
			job.latestRunPass = prowResult.Passed
			if *debug {
				log.Printf("Job: %s -- Latest Run No: %d -- Passed? %t", job, latestRunNo, prowResult.Passed)
			}
			if !prowResult.Passed {
				newFailures[job] = true
			}
		}

		if !job.latestRunPass {
			failedCount++
		}
	}
	return newFailures, failedCount
}

func getProwResult(target string) (*ProwResult, error) {
	jobFinishedFile := filepath.Join(target, finishedJSON)
	prowResultString, err := gcsClient.GetFileFromGCSString(*gcsBucket, jobFinishedFile)
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

// Add "do-not-merge/post-submit" labels to all PRs in protected repo towards protected branch
func blockPRs() {
	options := github.PullRequestListOptions{
		State: "open",
		Base:  *protectedBranch,
	}

	log.Printf("Adding [%s] to PRs in %s", doNotMergeLabel, *protectedRepo)
	if err := githubClnt.AddLabelToPRs(options, *protectedRepo, doNotMergeLabel); err != nil {
		log.Printf("Failed to add label to PRs: %v", err)
		return
	}
	log.Printf("Blocked auto-merge in %s, base: %s", *protectedRepo, *protectedBranch)
}

// remove "do-not-merge/post-submit" labels to all PRs in protected repo towards protected branch
func unBlockPRs() {
	options := github.PullRequestListOptions{
		State: "open",
		Base:  *protectedBranch,
	}

	if err := githubClnt.RemoveLabelFromPRs(options, *protectedRepo, doNotMergeLabel); err != nil {
		log.Printf("Failed to remove label to PRs: %v", err)
		return
	}
	log.Printf("PRs are clear to be auto-merged in %s, base: %s", *protectedRepo, *protectedBranch)
}

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

const (
	// TODO: Read from config file
	// Message Info
	sender          = "istio.testing@gmail.com"
	oncallMaillist  = "istio-oncall@googlegroups.com"
	adminMaillist   = "yutongz@google.com"
	messageSubject  = "[EMERGENCY] istio Post Submit failed!"
	messagePrologue = "Hi istio-oncall,\n\n" +
		"Post-Submit is failing in istio/istio, please take a look at following failure(s) and fix ASAP\n\n"
	messageEnding = "\nIf you have any questions about this message or notice inaccuracy, please contact istio-engprod@google.com."

	// Gmail setting
	gmailSMTPSERVER = "smtp.gmail.com"
	gmailSMTPPORT   = 587

	// Prow result GCS
	lastBuildTXT  = "latest-build.txt"
	finishedJSON  = "finished.json"
	gubernatorURL = "https://k8s-gubernator.appspot.com/build/istio-prow"

	doNotMergeLabel = "do-not-merge/post-submit"

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

	// Record which test run we already checked, avoid sending multiple email for the same test failure.
	bookkeeper map[string]int

	// TODO: Read from config file
	protectedPostsubmits = []string{"istio-postsubmit", "e2e-suite-rbac-auth", "e2e-suite-rbac-no_auth"}
	protectedRepo        = "istio"
	protectedBranch      = "master"
	receivers            = []string{oncallMaillist}
	gmailAppPass         string
)

func init() {
	flag.Parse()

	var err error
	gcsClient = u.NewGCSClient()

	token, err := u.GetAPITokenFromFile(*tokenFile)
	if err != nil {
		log.Fatalf("Error accessing user supplied token_file: %v\n", err)
	}
	githubClnt = u.NewGithubClient(*owner, token)

	if gmailAppPass, err = u.GetPasswordFromFile(*gmailAppPassFile); err != nil {
		log.Fatalf("Error accessing gmail app password: %v", err)
	}

	bookkeeper = make(map[string]int)
}

func main() {
	for {
		failures := getPostSubmitStatus()
		if len(failures) > 0 {
			log.Printf("%d tests failed in last circle", len(failures))
			sendMessage(formatMessage(failures))
			blockPRs()
		} else {
			log.Printf("No new tests failed in last circle.")
			unBlockPRs()
		}
		log.Printf("Sleeping for %d minutes", *interval)
		time.Sleep(time.Duration(*interval) * time.Minute)
	}
}

func formatMessage(failures map[string]bool) (mess string) {
	for job := range failures {
		mess += fmt.Sprintf("%s failed: %s/%s/%d\n\n", job, gubernatorURL, job, bookkeeper[job])
	}
	return
}

// Use gmail smtp server to send out email.
func sendMessage(body string) {
	if *debug {
		receivers = append(receivers, adminMaillist)
	}
	msg := fmt.Sprintf("From: %s\n", sender) +
		fmt.Sprintf("To: %s\n", receivers) +
		fmt.Sprintf("Subject: %s [%s]\n\n", messageSubject, time.Now().String()) +
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
// Record failure if it hasn't been recorded and update bookkeeper if necessary
func getPostSubmitStatus() map[string]bool {
	// Use this as a set, if a job failed, set true, otherwise not put in
	// So if no tracked job failed, this map should be empty
	failures := make(map[string]bool)

	for _, job := range protectedPostsubmits {
		latestRunNo, err := getLatestRun(job)
		if err != nil {
			log.Printf("Failed to get last run number of %s: %v", job, err)
			continue
		}

		if *debug {
			log.Printf("Job: %s -- Latest Run No: %d -- Recorded Run No: %d", job, latestRunNo, bookkeeper[job])
		}
		// No new test finished for this job
		if latestRunNo <= bookkeeper[job] {
			continue
		}

		prowResult, err := getProwResult(filepath.Join(job, strconv.Itoa(latestRunNo)))
		if err != nil {
			log.Printf("Failed to get prowResult %s/%d", job, latestRunNo)
			continue
		}
		bookkeeper[job] = latestRunNo
		if *debug {
			log.Printf("Job: %s -- Latest Run No: %d -- Passed? %t", job, latestRunNo, prowResult.Passed)
		}
		if !prowResult.Passed {
			failures[job] = true
		}
	}
	return failures
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
		Base:  protectedBranch,
	}

	log.Printf("Adding [%s] to PRs in %s", doNotMergeLabel, protectedRepo)
	if err := githubClnt.AddLabelToPRs(options, protectedRepo, doNotMergeLabel); err != nil {
		log.Printf("Failed to add label to PRs: %v", err)
		return
	}
	log.Printf("Blocked auto-merge in %s, base: %s", protectedRepo, protectedBranch)
}

// remove "do-not-merge/post-submit" labels to all PRs in protected repo towards protected branch
func unBlockPRs() {
	options := github.PullRequestListOptions{
		State: "open",
		Base:  protectedBranch,
	}

	log.Printf("Removing any [%s] from PRs in %s, base: %s", doNotMergeLabel, protectedRepo, protectedBranch)
	if err := githubClnt.RemoveLabelFromPRs(options, protectedRepo, doNotMergeLabel); err != nil {
		log.Printf("Failed to remove label to PRs: %v", err)
		return
	}
	log.Printf("PRs are clear to be auto-merged in %s, base: %s", protectedRepo, protectedBranch)
}

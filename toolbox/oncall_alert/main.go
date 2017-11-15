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
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"path/filepath"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
)

type ProwResult struct {
	TimeStamp  uint32                 `json:"timestamp"`
	Version    string                 `json:"version"`
	Result     string                 `json:"result"`
	Passed     bool                   `json:"passed"`
	JobVersion string                 `json:"job-version"`
	Metadata   map[string]interface{} `json:"metadata"`
}

const (
	sender          = "istio.testing@gmail.com"
	oncallMaillist  = "istio-oncall@googlegroups.com"
	messageSubject  = "[EMERGENCY] istio post-submit failed!"
	messagePrologue = "Hi istio-oncall,\n\n" +
		"Post-Submit is failing in istio/istio, please take a look at following failure(s) and fix ASAP\n\n"
	messageEnding = "\nIf you have any questions about this message or notice inaccuracy, please contact istio-engprod@google.com."

	gmailSMTPSERVER = "smtp.gmail.com"
	gmailSMTPPORT   = 587

	lastBuildTXT  = "latest-build.txt"
	finishedJson  = "finished.json"
	gubernatorURL = "https://k8s-gubernator.appspot.com/build/istio-prow"
)

var (
	gcsClient *storage.Client

	gcsBucket = flag.String("bucket", "istio-prow", "Prow artifact GCS bucket name.")
	interval  = flag.Int("interval", 5, "Check and report interval(minute)")
	debug     = flag.Bool("debug", false, "Optional to log debug message")

	bookkeeper map[string]int

	trackedPostsubmits = []string{"istio-postsubmit", "e2e-suite-rbac-auth", "e2e-suite-rbac-no_auth"}
)

func init() {
	flag.Parse()
	ctx := context.Background()

	var err error
	gcsClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create a gcs client, %v", err)
	}

	bookkeeper = make(map[string]int)
}

func main() {
	for {
		failures := getPostSubmitStatus()
		if len(failures) > 0 {
			log.Printf("%d tests failed in last circle", len(failures))
			message := FormatMessage(failures)
			sendMessage(message)
		} else {
			log.Printf("No new tests failed in last circle.")
		}
		time.Sleep(time.Duration(*interval) * time.Minute)
	}
}

func FormatMessage(failures map[string]bool) (mess string) {
	for job := range failures {
		mess += fmt.Sprintf("%s failed: %s/%s/%d\n\n", job, gubernatorURL, job, bookkeeper[job])
	}
	return
}

func sendMessage(body string) {
	appPass := "fnmhrqpnwlmpblfg"

	msg := fmt.Sprintf("From: %s\n", sender) +
		fmt.Sprintf("To: %s\n", oncallMaillist) +
		fmt.Sprintf("Subject: %s\n\n", messageSubject) +
		messagePrologue + body + messageEnding

	gmailSMTPAddr := fmt.Sprintf("%s:%d", gmailSMTPSERVER, gmailSMTPPORT)
	err := smtp.SendMail(gmailSMTPAddr, smtp.PlainAuth("istio-bot", sender, appPass, gmailSMTPSERVER),
		sender, []string{oncallMaillist}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Print("Message sent!")
}

func getLatestRun(job string) (int, error) {
	lastBuildFile := filepath.Join(job, lastBuildTXT)
	latestBuildString, err := getFileFromGCSString(*gcsBucket, lastBuildFile)
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

func getPostSubmitStatus() map[string]bool {
	// Use this as a set, if a job failed, set true, otherwise not put in
	// So if no tracked job failed, this map should be empty
	failures := make(map[string]bool)

	for _, job := range trackedPostsubmits {
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
	jobFinishedFile := filepath.Join(target, finishedJson)
	prowResultString, err := getFileFromGCSString(*gcsBucket, jobFinishedFile)
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

// Caller is responsible to close reader afterwards.
func getFileFromGCSReader(bucket, obj string) (*storage.Reader, error) {
	ctx := context.Background()
	r, err := gcsClient.Bucket(bucket).Object(obj).NewReader(ctx)

	if err != nil {
		log.Printf("Failed to download file %s/%s from gcs, %v", bucket, obj, err)
		return nil, err
	}

	return r, nil
}

func getFileFromGCSString(bucket, obj string) (string, error) {
	r, err := getFileFromGCSReader(bucket, obj)
	if err != nil {
		return "", err
	}
	defer func() {
		if err = r.Close(); err != nil {
			log.Printf("Failed to close gcs file reader, %v", err)
		}
	}()

	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(r); err != nil {
		log.Printf("Failed to read from gcs reader, %v", err)
		return "", err
	}

	return buf.String(), nil
}

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
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"

	u "istio.io/test-infra/toolbox/util"
)

// CI defines common accessors for continuous integration systems
type CI interface {
	GetLatestRun(jobName string) (int, error)
	GetResult(jobName string, runNo int) (*Result, error)
	Rerun(jobName string, runNo, numRerun int) error
	GetDetailsURL(jobName string, runNo int) string
}

// Result defines the output of each CI run
type Result struct {
	Passed bool   `json:"passed"`
	SHA    string `json:"sha"`
}

const (
	lastBuildTXT = "latest-build.txt"
	finishedJSON = "finished.json"
	startedJSON  = "started.json"
)

// ProwResult matches the structure published in finished.json
type ProwResult struct {
	TimeStamp  int64        `json:"timestamp"`
	Version    string       `json:"version"`
	Result     string       `json:"result"`
	Passed     bool         `json:"passed"`
	JobVersion string       `json:"job-version"`
	Metadata   ProwMetadata `json:"metadata"`
}

// ProwMetadata matches the structure published in finished.json
type ProwMetadata struct {
	Repo       string `json:"repo"`
	RepoCommit string `json:"repo-commit"`
}

// ProwJobConfig matches the structure published in started.json
type ProwJobConfig struct {
	Node        string            `json:"node"`
	JenkinsNode string            `json:"jenkins-node"`
	Version     string            `json:"version"`
	TimeStamp   int64             `json:"timestamp"`
	RepoVersion string            `json:"repo-version"`
	Pull        string            `json:"pull"`
	Repos       map[string]string `json:repos`
}

// ProwAccessor provides programmable access to Prow data on GCS
type ProwAccessor struct {
	prowProject   string
	prowZone      string
	gubernatorURL string
	gcsClient     u.IGCSClient
	rerunCmd      func(node string) error
}

// NewProwAccessor creates a new ProwAccessor
func NewProwAccessor(prowProject, prowZone, gubernatorURL, gcsBucket string) *ProwAccessor {
	return &ProwAccessor{
		prowProject:   prowProject,
		prowZone:      prowZone,
		gubernatorURL: gubernatorURL,
		gcsClient:     u.NewGCSClient(gcsBucket),
		rerunCmd: func(node string) error {
			_, e := u.Shell("kubectl create -f \"https://prow.istio.io/rerun?prowjob=%s\"", node)
			return e
		},
	}
}

// GetLatestRun reads latest run from "latest-build.txt" of a job under the gcs bucket
// Example: {gcsBucket}/{jobName}/latest-build.txt
func (p *ProwAccessor) GetLatestRun(jobName string) (int, error) {
	lastBuildFile := filepath.Join(jobName, lastBuildTXT)
	latestBuildString, err := p.gcsClient.Read(lastBuildFile)
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

// GetResult returns the Result of the job at a specific run
func (p *ProwAccessor) GetResult(jobName string, runNo int) (*Result, error) {
	jobFinishedFile := filepath.Join(jobName, strconv.Itoa(runNo), finishedJSON)
	prowResultString, err := p.gcsClient.Read(jobFinishedFile)
	if err != nil {
		log.Printf("Cannot access %s on GCS: %v", jobFinishedFile, err)
		return nil, err
	}
	prowResult := ProwResult{}
	if err = json.Unmarshal([]byte(prowResultString), &prowResult); err != nil {
		log.Printf("Failed to unmarshal ProwResult %s: %v", prowResultString, err)
		return nil, err
	}
	return &Result{
		Passed: prowResult.Passed,
		SHA:    prowResult.Metadata.RepoCommit,
	}, nil
}

// GetDetailsURL returns the gubernator URL to that job at the run number
func (p *ProwAccessor) GetDetailsURL(jobName string, runNo int) string {
	return fmt.Sprintf("%s/%s/%d", p.gubernatorURL, jobName, runNo)
}

// Rerun starts on Prow the reruns on specified jobs
func (p *ProwAccessor) Rerun(jobName string, runNo, numRerun int) error {
	cfg, err := p.getProwJobConfig(jobName, runNo)
	if err != nil {
		return err
	}
	if err = p.triggerConcurrentReruns(jobName, cfg.Node, numRerun); err != nil {
		return err
	}
	return nil
}

// getProwJobConfig fetches the config of the job at runNo
func (p *ProwAccessor) getProwJobConfig(jobName string, runNo int) (*ProwJobConfig, error) {
	jobStartedFile := filepath.Join(jobName, strconv.Itoa(runNo), startedJSON)
	StartedFileString, err := p.gcsClient.Read(jobStartedFile)
	if err != nil {
		return nil, err
	}
	cfg := ProwJobConfig{}
	if err = json.Unmarshal([]byte(StartedFileString), &cfg); err != nil {
		log.Printf("Failed to unmarshal ProwJobConfig %s: %v\n", StartedFileString, err)
		return nil, err
	}
	return &cfg, nil
}

func (p *ProwAccessor) triggerConcurrentReruns(jobName, node string, numRerun int) error {
	log.Printf("Rerunning %s\n", jobName)
	recess := 1 * time.Minute
	maxRetry := 3
	for i := 0; i < numRerun; i++ {
		if err := u.Retry(recess, maxRetry, func() error {
			return p.rerunCmd(node)
		}); err != nil {
			log.Printf("Unable to trigger the %d-th rerun of job %v", i, jobName)
		}
	}
	return nil
}

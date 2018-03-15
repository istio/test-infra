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
	"log"
	"path/filepath"
	"strconv"

	u "istio.io/test-infra/toolbox/util"
)

const (
	lastBuildTXT = "latest-build.txt"
	finishedJSON = "finished.json"
	startedJSON  = "started.json"
)

type IProwAccessor interface {
	GetLatestRun(jobName string) (int, error)
	GetProwResult(jobName string, runNo int) (*ProwResult, error)
	GetProwJobConfig(jobName string, runNo int) (*ProwJobConfig, error)
	GetGubernatorURL() string
}

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
	Repo       string `json:"repo"`
	RepoCommit string `json:"repo-commit"`
}

// ProwJobConfig matches the structure published in started.json
type ProwJobConfig struct {
	Node        string `json:"node"`
	JenkinsNode string `json:"jenkins-node"`
	Version     string `json:"version"`
	TimeStamp   uint32 `json:"timestamp"`
	RepoVersion string `json:"repo-version"`
}

// ProwAccessor provides programmable access to Prow data on GCS
type ProwAccessor struct {
	prowProject   string
	prowZone      string
	gubernatorURL string
	gcsClient     *u.GCSClient
}

// NewProwAccessor creates a new ProwAccessor
func NewProwAccessor(prowProject, prowZone, gubernatorURL, gcsBucket string) *ProwAccessor {
	return &ProwAccessor{
		prowProject:   prowProject,
		prowZone:      prowZone,
		gubernatorURL: gubernatorURL,
		gcsClient:     u.NewGCSClient(gcsBucket),
	}
}

// GetLatestRun reads latest run from "latest-build.txt" of a job under the gcs bucket
// Example: {gcsBucket}}/{jobName}/latest-build.txt
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

// GetProwResult returns the ProwResult of the job at a specific run
func (p *ProwAccessor) GetProwResult(jobName string, runNo int) (*ProwResult, error) {
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
	return &prowResult, nil
}

// GetProwJobConfig fetches the config of the job at runNo
func (p *ProwAccessor) GetProwJobConfig(jobName string, runNo int) (*ProwJobConfig, error) {
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

// GetGubernatorURL returns the gubernator URL used by this ProwAccessor
func (p *ProwAccessor) GetGubernatorURL() string {
	return p.gubernatorURL
}

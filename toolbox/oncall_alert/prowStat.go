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

package main

import (
	"encoding/json"
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

// ProwJobConfig matches the structure published in started.json
type ProwJobConfig struct {
	Node        string `json:"node"`
	JenkinsNode string `json:"jenkins-node"`
	Version     string `json:"version"`
	TimeStamp   uint32 `json:"timestamp"`
	RepoVersion string `json:"repo-version"`
}

// FlakeStat records the stats from flakiness detection by multiple reruns
type FlakeStat struct {
	TestName           string           `json:"testName"`
	SHA                string           `json:"sha"`
	TotalRerun         int              `json:"totalRerun"`
	Failures           int              `json:"failures"`
	ParentJobTimeStamp uint32           `json:"parentJobTimeStamp"`
	failedTestCases    []FailedTestCase `json:"failedTestCases"`
}

type FailedTestCase struct {
	Name       string `json:"name"`
	TotalRerun int    `json:"totalRerun"`
	Failures   int    `json:"failures"`
}

// SerializeFlakeStat flattens in-memory FlakeStat to string
func SerializeFlakeStat(flakeStat FlakeStat) (string, error) {
	pickled, err := json.MarshalIndent(flakeStat, "", "\t")
	return string(pickled), err
}

// SerializeFlakeStats flatten in-memory []FlakeStat to string
func SerializeFlakeStats(flakeStats []FlakeStat) (string, error) {
	pickled, err := json.MarshalIndent(flakeStats, "", "\t")
	return string(pickled), err
}

// DeserializeFlakeStats inflate the list of FlakeStat from string
func DeserializeFlakeStats(pickled string) ([]FlakeStat, error) {
	var flakeStats []FlakeStat
	err := json.Unmarshal([]byte(pickled), &flakeStats)
	return flakeStats, err
}

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

package ci_to_gubernator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	s "istio.io/test-infra/sisyphus"
	u "istio.io/test-infra/toolbox/util"
)

const (
	lastBuildTXT  = "latest-build.txt" // TODO update this file
	finishedJSON  = "finished.json"
	startedJSON   = "started.json"
	unknown       = "unknown"
	resultSuccess = "SUCCESS"
	resultFailure = "FAILURE"
)

type Converter struct {
	gcsClient u.IGCSClient
	org       string
	repo      string
	job       string
	build     int
}

func NewConverter(bucket, org, repo, job string, build int) *Converter {
	return &Converter{
		gcsClient: u.NewGCSClient(bucket),
		org:       org,
		repo:      repo,
		job:       job,
		build:     build,
	}
}

func (c *Converter) CreateFinishedJSON(exitCode int, sha string) error {
	result := resultSuccess
	passed := true
	if exitCode != 0 {
		result = resultFailure
		passed = false
	}
	finished := s.ProwResult{
		TimeStamp:  time.Now().Unix(),
		Version:    unknown,
		Result:     result,
		Passed:     passed,
		JobVersion: unknown,
		Metadata: s.ProwMetadata{
			Repo:       fmt.Sprintf("github.com/%s/%s", c.org, c.repo),
			RepoCommit: sha,
		},
	}
	flattened, err := json.MarshalIndent(finished, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(finishedJSON, flattened, 0600)
}

// GenerateStartedJSON creates the string content of start.json
func (c *Converter) GenerateStartedJSON(prNum int, sha string) (string, error) {
	prNumColonSHA := fmt.Sprintf("%s:%s", prNum, sha)
	started := s.ProwJobConfig{
		TimeStamp: time.Now().Unix(),
		Pull:      prNumColonSHA,
		Repos: map[string]string{
			fmt.Sprintf("github.com/%s/%s", c.org, c.repo): prNumColonSHA,
		},
	}
	flattened, err := json.MarshalIndent(started, "", "\t")
	return string(flattened), err
}

// CreateUploadStartedJSON creates and uploads started.json
func (c *Converter) CreateUploadStartedJSON(prNum int, sha string) error {
	gcsPath := filepath.Join(c.job, string(c.build), startedJSON)
	return c.CreateUploadStartedJSONCustomPath(prNum, sha, gcsPath)
}

// CreateUploadStartedJSON creates and uploads started.json
func (c *Converter) CreateUploadStartedJSONCustomPath(prNum int, sha, gcsPath string) error {
	str, err := c.GenerateStartedJSON(prNum, sha)
	if err != nil {
		return err
	}
	return c.gcsClient.Write(gcsPath, str)
}

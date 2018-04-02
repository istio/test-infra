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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"

	"github.com/marcacohen/gcslock"

	s "istio.io/test-infra/sisyphus"
	u "istio.io/test-infra/toolbox/util"
)

const (
	lastBuildTXT  = "latest-build.txt" // TODO update this file
	buildLogTXT   = "build-log.txt"
	finishedJSON  = "finished.json"
	startedJSON   = "started.json"
	junitXML      = "junit.xml"
	artifacts     = "artifacts"
	unknown       = "unknown"
	resultSuccess = "SUCCESS"
	resultFailure = "FAILURE"
)

var (
	defaultTimeout = 5 * time.Minute
)

// Converter defines how to convert generic CI results to artifacts that gubernator understands
type Converter struct {
	gcsClient     u.IGCSClient
	bucket        string
	gcsPathPrefix string
	org           string
	repo          string
	job           string
	build         int
}

// NewConverter creates a Converter
func NewConverter(bucket, org, repo, job string, build int) *Converter {
	return &Converter{
		gcsClient:     u.NewGCSClient(bucket),
		gcsPathPrefix: fmt.Sprintf("%s/%d", job, build),
		bucket:        bucket,
		org:           org,
		repo:          repo,
		job:           job,
		build:         build,
	}
}

// SetGCSPathPrefix allows customized gcs location
func (c *Converter) SetGCSPathPrefix(prefix string) *Converter {
	c.gcsPathPrefix = prefix
	return c
}

// GenerateFinishedJSON creates the string content of finished.json
func (c *Converter) GenerateFinishedJSON(exitCode int, sha string) (string, error) {
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
	return string(flattened), err
}

// CreateUploadFinishedJSON creates and uploads finished.json
func (c *Converter) CreateUploadFinishedJSON(exitCode int, sha string) error {
	log.Printf("Generating finished.json")
	str, err := c.GenerateFinishedJSON(exitCode, sha)
	if err != nil {
		return err
	}
	gcsPath := filepath.Join(c.gcsPathPrefix, finishedJSON)
	log.Printf("Uploading finished.json to gs://%s/%s", c.bucket, gcsPath)
	return c.gcsClient.Write(gcsPath, str)
}

// UploadJunitReports uploads junit report to GCS
func (c *Converter) UploadJunitReports(junitReport string) error {
	gcsPath := filepath.Join(c.gcsPathPrefix, artifacts, junitXML)
	return c.upload(junitReport, gcsPath)
}

// UploadBuildLog uploads build-log.txt to GCS
func (c *Converter) UploadBuildLog(logFile string) error {
	gcsPath := filepath.Join(c.gcsPathPrefix, buildLogTXT)
	return c.upload(logFile, gcsPath)
}

// UpdateLastBuildTXT updates latest-build.txt to GCS
func (c *Converter) UpdateLastBuildTXT() error {
	m, err := gcslock.New(nil, c.bucket, lastBuildTXT)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err = m.ContextLock(ctx); err != nil {
		return err
	}
	defer func() {
		if err = m.ContextUnlock(ctx); err != nil {
			log.Fatalf("Unlock %s has timed out: %v", lastBuildTXT, err)
		}
	}()
	gcsPath := filepath.Join(c.job, lastBuildTXT)
	val, err := c.gcsClient.Read(gcsPath)
	if err != nil {
		return err
	}
	log.Printf("Current value of gs://%s is %s", gcsPath, val)
	latestBuildInt, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	if c.build > latestBuildInt {
		log.Printf("Updating gs://%s to be %d", gcsPath, c.build)
		return c.gcsClient.Write(gcsPath, fmt.Sprintf("%d", c.build))
	}
	log.Printf("No updates on %s needed", gcsPath)
	return nil
}

// GenerateStartedJSON creates the string content of start.json
func (c *Converter) GenerateStartedJSON(prNum int, sha string) (string, error) {
	prNumColonSHA := fmt.Sprintf("%d:%s", prNum, sha)
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
	log.Printf("Generating started.json")
	str, err := c.GenerateStartedJSON(prNum, sha)
	if err != nil {
		return err
	}
	gcsPath := filepath.Join(c.gcsPathPrefix, startedJSON)
	log.Printf("Uploading started.json to gs://%s/%s", c.bucket, gcsPath)
	return c.gcsClient.Write(gcsPath, str)
}

func (c *Converter) upload(localPath, gcsPath string) error {
	log.Printf("Uploading %s to gs://%s/%s", localPath, c.bucket, gcsPath)
	str, err := u.ReadFile(localPath)
	if err != nil {
		return err
	}
	return c.gcsClient.Write(gcsPath, str)
}

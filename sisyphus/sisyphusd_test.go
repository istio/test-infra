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
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

const (
	prowProjectMock   = "prow-project-mock"
	prowZoneMock      = "us-west1-a"
	gubernatorURLMock = "https://k8s-gubernator.appspot.com/build/mock"
	gcsBucketMock     = "gcs-bucket-mock"
	testDataDir       = "test_data"
)

var (
	protectedJobsMock = []string{"job-1" /*, "job-2", "job-3"*/}
)

type ProwAccessorMock struct {
	lastestRuns   map[string]int // jobName -> lastestRun
	gubernatorURL string
	prowResults   map[string]map[int]ProwResult // jobName -> runNo -> ProwResult
}

func NewProwAccessorMock(gubernatorURL string) *ProwAccessorMock {
	prowResults := map[string]map[int]ProwResult{}
	for _, job := range protectedJobsMock {
		dataFile := filepath.Join(testDataDir, job+".json")
		raw, err := ioutil.ReadFile(dataFile)
		if err != nil {
			log.Fatalf("Error reading %s:%v", dataFile, err)
		}
		// intermediate step since json does not support integer as key
		prowResultsForOneJobTmp := map[string]ProwResult{}
		if err = json.Unmarshal([]byte(raw), &prowResultsForOneJobTmp); err != nil {
			log.Fatalf("Failed to unmarshal test data for %s: %v", dataFile, err)
		}
		prowResultsForOneJob := map[int]ProwResult{}
		for k, v := range prowResultsForOneJobTmp {
			i, err := strconv.Atoi(k)
			if err != nil {
				log.Fatalf("Error converting %s to int :%v", k, err)
			}
			prowResultsForOneJob[i] = v
		}
		prowResults[job] = prowResultsForOneJob
	}
	return &ProwAccessorMock{
		lastestRuns:   make(map[string]int),
		gubernatorURL: gubernatorURL,
		prowResults:   prowResults,
	}
}

func (p *ProwAccessorMock) GetLatestRun(jobName string) (int, error) {
	p.lastestRuns[jobName]++
	return p.lastestRuns[jobName], nil
}

func (p *ProwAccessorMock) GetProwResult(jobName string, runNo int) (*ProwResult, error) {
	res := p.prowResults[jobName][runNo]
	return &res, nil
}

func (p *ProwAccessorMock) GetProwJobConfig(jobName string, runNo int) (*ProwJobConfig, error) {
	return nil, nil
}

func (p *ProwAccessorMock) GetGubernatorURL() string {
	return p.gubernatorURL
}

func (p *ProwAccessorMock) Rerun(jobName string, runNo, numRerun int) error {
	return nil
}

func TestSisyphusDaemonConfig(t *testing.T) {
	catchFlakesByRun := true
	cfg := &SisyphusConfig{
		CatchFlakesByRun: catchFlakesByRun,
	}
	cfgExpected := &SisyphusConfig{
		CatchFlakesByRun: catchFlakesByRun,
		PollGapSecs:      DefaultPollGapSecs,
		NumRerun:         DefaultNumRerun,
	}
	sd := SisyphusDaemon(protectedJobsMock, prowProjectMock,
		prowZoneMock, gubernatorURLMock, gcsBucketMock, cfg)
	if !reflect.DeepEqual(sd.GetSisyphusConfig(), cfgExpected) {
		t.Error("setting catchFlakesByRun failed")
	}
}

func TestProwResultsMockParsing(t *testing.T) {
	prowAccessor := NewProwAccessorMock(gubernatorURLMock)
	res, err := prowAccessor.GetProwResult("job-1", 10)
	if err != nil {
		t.Error("GetProwResult failed: %v", err)
	}
	if res.Metadata.RepoCommit != "sha-1" {
		t.Error("RepoCommit unmatched with data in file")
	}
}

func TestRerunLogics(t *testing.T) {

}

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
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"
)

const (
	prowProjectMock        = "prow-project-mock"
	prowZoneMock           = "us-west1-a"
	gubernatorURLMock      = "https://k8s-gubernator.appspot.com/build/mock"
	testDataDir            = "test_data"
	expectedFlakeStatsJSON = "expectedFlakeStats.json"
)

var (
	protectedJobsMock = []string{"job-1" /*, "job-2", "job-3"*/}
)

type ProwAccessorMock struct {
	lastestRunNos   map[string]int // jobName -> lastestRun
	maxRunNos       map[string]int // jobName -> maxRunNo
	gubernatorURL   string
	prowResults     map[string]map[int]ProwResult // jobName -> runNo -> ProwResult
	cancelSisyphusd context.CancelFunc
}

func NewProwAccessorMock(gubernatorURL string) *ProwAccessorMock {
	mock := &ProwAccessorMock{
		lastestRunNos:   make(map[string]int),
		maxRunNos:       make(map[string]int),
		gubernatorURL:   gubernatorURL,
		prowResults:     make(map[string]map[int]ProwResult),
		cancelSisyphusd: func() {},
	}
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
		minRunNo := 999999
		maxRunNo := 0
		for k, v := range prowResultsForOneJobTmp {
			i, err := strconv.Atoi(k)
			if err != nil {
				log.Fatalf("Error converting %s to int :%v", k, err)
			}
			if i > maxRunNo {
				maxRunNo = i
			}
			if i < minRunNo {
				minRunNo = i
			}
			prowResultsForOneJob[i] = v
		}
		mock.prowResults[job] = prowResultsForOneJob
		mock.lastestRunNos[job] = minRunNo
		mock.maxRunNos[job] = maxRunNo
	}
	return mock
}

func (p *ProwAccessorMock) GetLatestRun(jobName string) (int, error) {
	ret := p.lastestRunNos[jobName]
	p.lastestRunNos[jobName]++
	if ret > p.maxRunNos[jobName] {
		allRunsFinished := true
		for j, lastestRun := range p.lastestRunNos {
			if lastestRun <= p.maxRunNos[j] { // equal sign to ensure the last run is seen before termination
				allRunsFinished = false
			}
		}
		if allRunsFinished {
			p.cancelSisyphusd()
		}
		return p.maxRunNos[jobName], nil
	}
	return ret, nil
}

func (p *ProwAccessorMock) GetResult(jobName string, runNo int) (*Result, error) {
	prowResult := p.prowResults[jobName][runNo]
	return &Result{
		Passed: prowResult.Passed,
		SHA:    prowResult.Metadata.RepoCommit,
	}, nil
}

func (p *ProwAccessorMock) GetDetailsURL(jobName string, runNo int) string {
	return ""
}

func (p *ProwAccessorMock) Rerun(jobName string, runNo, numRerun int) error {
	return nil
}

type StorageMock struct {
	t             *testing.T
	expectedStats []FlakeStat
}

func NewStorageMock(t *testing.T) *StorageMock {
	dataFile := filepath.Join(testDataDir, expectedFlakeStatsJSON)
	raw, err := ioutil.ReadFile(dataFile)
	if err != nil {
		log.Fatalf("Error reading %s:%v", dataFile, err)
	}
	var expectedStats []FlakeStat
	if err = json.Unmarshal([]byte(raw), &expectedStats); err != nil {
		log.Fatalf("Failed to unmarshal test data for %s: %v", dataFile, err)
	}
	return &StorageMock{
		t:             t,
		expectedStats: expectedStats,
	}
}

func (s *StorageMock) Store(jobName, sha string, newFlakeStat FlakeStat) error {
	for _, expectedStat := range s.expectedStats {
		if expectedStat.TestName == jobName && expectedStat.SHA == sha {
			if !reflect.DeepEqual(expectedStat, newFlakeStat) {
				s.t.Errorf("Expecting %v but got %v", expectedStat, newFlakeStat)
			}
			return nil
		}
	}
	s.t.Errorf("No matching expectedStat found given jobName = %s, sha = %s", jobName, sha)
	return nil
}

type fakeClient struct{}

func (f fakeClient) Read(obj string) (string, error) {
	return "data", nil
}

func (f fakeClient) Write(obj, txt string) error {
	return nil
}

func TestDaemonConfig(t *testing.T) {
	catchFlakesByRun := true
	cfg := &Config{
		CatchFlakesByRun: catchFlakesByRun,
	}
	cfgExpected := &Config{
		CatchFlakesByRun: catchFlakesByRun,
		PollGapDuration:  DefaultPollGapDuration,
		NumRerun:         DefaultNumRerun,
	}
	presubmitJobs := []string{}
	sisyphusd := NewDaemonUsingProw(
		protectedJobsMock,
		presubmitJobs,
		prowProjectMock, prowZoneMock, gubernatorURLMock,
		gcsBucketMock,
		fakeClient{},
		NewStorageMock(t), cfg)
	if !reflect.DeepEqual(sisyphusd.GetConfig(), cfgExpected) {
		t.Error("setting catchFlakesByRun failed")
	}
}

func TestProwResultsMock(t *testing.T) {
	job := "job-1"
	prowAccessorMock := NewProwAccessorMock(gubernatorURLMock)
	res, err := prowAccessorMock.GetResult(job, 10)
	if err != nil {
		t.Errorf("GetProwResult failed: %v", err)
	}
	if res.SHA != "sha-1" {
		t.Error("RepoCommit unmatched with data in file")
	}
	expectedBase := 10
	if runNo, _ := prowAccessorMock.GetLatestRun(job); runNo != expectedBase {
		t.Errorf("Expecting frist call to GetLatestRun to return %d but got %d", expectedBase, runNo)
	}
	if runNo, _ := prowAccessorMock.GetLatestRun(job); runNo != expectedBase+1 {
		t.Errorf("Expecting second call to GetLatestRun to return %d but got %d", expectedBase+1, runNo)
	}
	for i := 0; i < 100; i++ {
		prowAccessorMock.GetLatestRun(job)
	}
	expectedMax := 15
	if runNo, _ := prowAccessorMock.GetLatestRun(job); runNo != expectedMax {
		t.Errorf("Expecting call to GetLatestRun to not exceed %d but got %d", expectedMax, runNo)
	}
}

func TestRerunLogics(t *testing.T) {
	sisyphusd := newDaemon(
		protectedJobsMock,
		&Config{
			CatchFlakesByRun: true,
			PollGapDuration:  100 * time.Millisecond,
		},
		NewStorageMock(t))
	prowAccessorMock := NewProwAccessorMock(gubernatorURLMock)
	ctx, cancelFn := context.WithCancel(context.Background())
	prowAccessorMock.cancelSisyphusd = cancelFn
	sisyphusd.ci = prowAccessorMock
	sisyphusd.Start(ctx)
}

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	os.Exit(m.Run())
}

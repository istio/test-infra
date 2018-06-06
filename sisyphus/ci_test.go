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
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	u "istio.io/test-infra/toolbox/util"
)

const (
	jobName          = "istio-postsubmit"
	runNo            = 1000
	finishedJSONMock = "mock-finished.json"
	startedJSONMock  = "mock-started.json"
	lastBuildTXTMock = "mock-latest-build.txt"
	nodeMock         = "206d0875-2405-11e8-b09d-0a580a2c2be7"
)

type gcsMock struct{}

func (gcs *gcsMock) Read(obj string) (string, error) {
	// always use the same mock
	var mock string
	if strings.Contains(obj, finishedJSON) {
		mock = finishedJSONMock
	} else if strings.Contains(obj, startedJSON) {
		mock = startedJSONMock
	} else if strings.Contains(obj, lastBuildTXT) {
		mock = lastBuildTXTMock
	} else {
		return "", fmt.Errorf("no mock found for %s", obj)
	}
	obj = filepath.Join("test_data", mock)
	return u.ReadFile(obj)
}

func (gcs *gcsMock) Write(obj, txt string) error {
	return nil
}

func (gcs *gcsMock) Exists(obj string) (bool, error) {
	return false, nil
}

func newProwAccessorWithGCSMock() *ProwAccessor {
	return &ProwAccessor{
		gcsClient: &gcsMock{},
		rerunCmd:  func(node string) error { return nil },
	}
}

func TestGetLatestRunOnProw(t *testing.T) {
	prowAccessor := newProwAccessorWithGCSMock()
	res, err := prowAccessor.GetLatestRun(jobName)
	if err != nil {
		t.Errorf("Error when calling GetLatestRun: %v", err)
	}
	expected := 123
	if !reflect.DeepEqual(res, expected) {
		t.Errorf("ProwResult retrieved from Prow is not as expected")
	}
}

func TestGetResultOnProw(t *testing.T) {
	prowAccessor := newProwAccessorWithGCSMock()
	res, err := prowAccessor.GetResult(jobName, runNo)
	if err != nil {
		t.Errorf("Error when calling GetResult: %v", err)
	}
	expectedResult := &Result{
		Passed: true,
		SHA:    "cc02e59987163c8deb80db8bcc203c318c1b752e",
	}
	if !reflect.DeepEqual(res, expectedResult) {
		t.Errorf("ProwResult retrieved from Prow is not as expected")
	}
}

func TestGetProwJobConfig(t *testing.T) {
	prowAccessor := newProwAccessorWithGCSMock()
	cfg, err := prowAccessor.getProwJobConfig(jobName, runNo)
	if err != nil {
		t.Errorf("Error when calling GetProwJobConfig: %v", err)
	}
	cfgExpected := &ProwJobConfig{
		Node:        nodeMock,
		JenkinsNode: nodeMock,
		Version:     "unknown",
		TimeStamp:   1520646538,
		RepoVersion: "unknown",
		Pull:        "some_pr_number:sha",
		Repos: map[string]string{
			"repo": "master:1920a01a8df75baf71ae5fa38d68e4b055bad65c,554:cc02e59987163c8deb80db8bcc203c318c1b752e"},
	}
	if !reflect.DeepEqual(cfg, cfgExpected) {
		fmt.Printf("cfg = %v\n", cfg)
		fmt.Printf("cfgExpected = %v\n", cfgExpected)
		t.Errorf("ProwJobConfig retrieved from Prow is not as expected")
	}
}

func TestRerunOnProw(t *testing.T) {
	prowAccessor := newProwAccessorWithGCSMock()
	counts := make(map[string]int)
	prowAccessor.rerunCmd = func(node string) error {
		counts[node]++
		return nil
	}
	expectedNumRerun := 1
	err := prowAccessor.Rerun(jobName, runNo)
	if err != nil {
		t.Errorf("Error when calling GetProwJobConfig: %v", err)
	}
	if counts[nodeMock] != expectedNumRerun {
		t.Errorf("Expect %d reruns but only did %d", expectedNumRerun, counts[nodeMock])
	}
}

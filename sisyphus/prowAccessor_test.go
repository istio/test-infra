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
	"reflect"
	"testing"
)

const (
	prowProject   = "istio-testing"
	prowZone      = "us-west1-a"
	gubernatorURL = "https://k8s-gubernator.appspot.com/build/istio-prow"
	gcsBucket     = "istio-prow"
	jobName       = "istio-postsubmit"
	runNo         = 1000
)

func TestGetProwResult(t *testing.T) {
	prowAccessor := NewProwAccessor(prowProject, prowZone, gubernatorURL, gcsBucket)
	prowResult, err := prowAccessor.GetProwResult(jobName, runNo)
	if err != nil {
		t.Errorf("Error when calling GetProwResult: %v", err)
	}
	prowResultExpected := &ProwResult{
		TimeStamp:  1519754919,
		Version:    "unknown",
		Result:     "SUCCESS",
		Passed:     true,
		JobVersion: "unknown",
		Metadata: ProwMetadata{
			Repo:       "github.com/istio/istio",
			RepoCommit: "5a3de490ffdfe39672c706e5086b8c6977299b01",
		},
	}
	if !reflect.DeepEqual(prowResult, prowResultExpected) {
		t.Errorf("ProwResult retrieved from Prow is not as expected")
	}
}

func TestGetProwJobConfig(t *testing.T) {
	prowAccessor := NewProwAccessor(prowProject, prowZone, gubernatorURL, gcsBucket)
	cfg, err := prowAccessor.getProwJobConfig(jobName, runNo)
	if err != nil {
		t.Errorf("Error when calling GetProwJobConfig: %v", err)
	}
	cfgExpected := &ProwJobConfig{
		Node:        "958adeee-1be5-11e8-b0d2-0a580a2c0603",
		JenkinsNode: "958adeee-1be5-11e8-b0d2-0a580a2c0603",
		Version:     "unknown",
		TimeStamp:   1519753382,
		RepoVersion: "unknown",
	}
	if !reflect.DeepEqual(cfg, cfgExpected) {
		t.Errorf("ProwJobConfig retrieved from Prow is not as expected")
	}
}

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
)

var (
	protectedJobs = []string{"istio-postsubmit", "e2e-suite-rbac-auth", "e2e-suite-rbac-no_auth"}
)

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
	sd := SisyphusDaemon(protectedJobs, prowProject, prowZone, gubernatorURL, gcsBucket, cfg)
	if !reflect.DeepEqual(sd.GetSisyphusConfig(), cfgExpected) {
		t.Error("setting catchFlakesByRun failed")
	}
}

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
// limitations under the License.package main

package main

import (
	"strings"
	"testing"
)

func TestFindMaster(t *testing.T) {
	testingInput := []byte(`
repos:
  istio:
    branches:
        <<: *blocked_branches
        master:
          protect: true
  b:
    branches:
        <<: *blocked_branches
        master:
          protect: true
          required_status_checks:
            contexts:
            - "ci/circleci: e2e-pilot-cloudfoundry-v1alpha3-v2"
`)

	output := findRepo(testingInput, []string{"istio"}, "a")
	correctOutput := []byte(`
repos:
  istio:
    branches:
        <<: *blocked_branches
        a:
          protect: true
          required_status_checks:
            contexts:
            - "merges-blocked-needs-admin"
        master:
          protect: true
  b:
    branches:
        <<: *blocked_branches
        master:
          protect: true
          required_status_checks:
            contexts:
            - "ci/circleci: e2e-pilot-cloudfoundry-v1alpha3-v2"
`)
	if strings.Compare(output, string(correctOutput)) != 0 {
		t.Fail()
	}

}

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
	"time"

	s "istio.io/test-infra/sisyphus"
)

const (
	lastBuildTXT  = "latest-build.txt" // TODO update this file
	finishedJSON  = "finished.json"
	startedJSON   = "started.json"
	unknown       = "unknown"
	resultSuccess = "SUCCESS"
	resultFailure = "FAILURE"
)

func CreateFinishedJSON(exitCode int, sha, org, repo string) error {
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
			Repo:       fmt.Sprintf("github.com/%s/%s", org, repo),
			RepoCommit: sha,
		},
	}
	flattened, err := json.MarshalIndent(finished, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(finishedJSON, flattened, 0600)
}

func CreateStartedJSON(prNum int /*, sha, org, repo string*/) error {
	// started := ProwJobConfig{
	// 	TimeStamp: time.Now().Unix(),
	// }
	// flattened, err := json.MarshalIndent(finished, "", "\t")
	// if err != nil {
	// 	return err
	// }
	// return ioutil.WriteFile(finishedJSON, flattened, 0600)
	return nil
}

// Copyright 2017 Istio Authors
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
	"io/ioutil"
)

/*
	Assumes Dependency Hoisting in all bazel dependency files. Example:

		new_git_repository(
			name = "mixerapi_git",
			commit = "ee9769f5b3304d9e01cd7ed6fb1dbb9b08e96210",
			remote = "https://github.com/istio/api.git",
		)

	becomes

		MIXERAPI = "ee9769f5b3304d9e01cd7ed6fb1dbb9b08e96210"
		new_git_repository(
			name = "mixerapi_git",
			commit = MIXERAPI,
			remote = "https://github.com/istio/api.git",
		)
*/

type dependency struct {
	Name       string `json:"name"`
	RepoName   string `json:"repoName"`
	ProdBranch string `json:"prodBranch"` // either master or stable
	File       string `json:"file"`       // where in the *parent* repo such dependecy is recorded
}

// Get the list of dependencies of a repo
func getDeps(depsFilePath string) ([]dependency, error) {
	var deps []dependency
	raw, err := ioutil.ReadFile(depsFilePath)
	if err != nil {
		return deps, err
	}
	err = json.Unmarshal(raw, &deps)
	return deps, err
}

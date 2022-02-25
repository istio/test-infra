// Copyright 2019 Istio Authors
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

package pkg

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"istio.io/test-infra/tools/prowgen/pkg/spec"
)

func TestGenerateConfig(t *testing.T) {
	bc := ReadBase(nil, "testdata/.base.yaml")
	cli := &Client{BaseConfig: *bc}
	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name: "simple",
		},
		{
			name: "matrix",
		},
		{
			name: "params",
		},
		{
			name:        "long-job-name",
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := fmt.Sprintf("testdata/%s.yaml", tt.name)
			jobs := cli.ReadJobsConfig(file)
			for _, branch := range jobs.Branches {
				output, err := cli.ConvertJobConfig(file, jobs, branch)
				if tt.expectError {
					if err == nil {
						t.Fatalf("Test %q expected an error, but did not receive one", tt.name)
					}
					// there should be no generated file when an error occurs
					continue
				} else if err != nil {
					t.Fatalf("Test %q did not expect an error, but received %v", tt.name, err)
				}
				testFile := fmt.Sprintf("testdata/%s.gen.yaml", tt.name)
				if os.Getenv("REFRESH_GOLDEN") == "true" {
					Write(output, testFile, bc.AutogenHeader)
				}
				if err := Check(output, testFile, bc.AutogenHeader); err != nil {
					t.Fatal(err.Error())
				}
			}
		})
	}
}

func TestFilterReleaseBranchingJobs(t *testing.T) {
	testCases := []struct {
		name         string
		jobs         []spec.Job
		filteredJobs []spec.Job
	}{
		{
			name:         "filter an empty list of jobs",
			jobs:         []spec.Job{},
			filteredJobs: []spec.Job{},
		},
		{
			name: "filter enabled release branching jobs",
			jobs: []spec.Job{
				{
					Name:    "job_1",
					Command: []string{"exit", "0"},
					Types:   []string{"presubmit"},
				},
				{
					Name:                    "job_2",
					Command:                 []string{"echo", "pass"},
					DisableReleaseBranching: false,
					Types:                   []string{"postsubmit"},
				},
			},
			filteredJobs: []spec.Job{
				{
					Name:    "job_1",
					Command: []string{"exit", "0"},
					Types:   []string{"presubmit"},
				},
				{
					Name:                    "job_2",
					Command:                 []string{"echo", "pass"},
					DisableReleaseBranching: false,
					Types:                   []string{"postsubmit"},
				},
			},
		},
		{
			name: "filter disabled release branching jobs",
			jobs: []spec.Job{
				{
					Name:                    "job_1",
					Command:                 []string{"exit", "0"},
					DisableReleaseBranching: true,
					Types:                   []string{"presubmit"},
				},
				{
					Name:                    "job_2",
					Command:                 []string{"echo", "pass"},
					DisableReleaseBranching: true,
					Types:                   []string{"postsubmit"},
				},
			},
			filteredJobs: []spec.Job{},
		},
	}

	for _, tc := range testCases {
		expected := tc.filteredJobs
		actual := FilterReleaseBranchingJobs(tc.jobs)

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Fatalf("Filtered jobs do not match, (-want, +got): \n%s", diff)
		}
	}
}

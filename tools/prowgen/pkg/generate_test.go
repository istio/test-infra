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
	"reflect"
	"testing"
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
			name: "simple-matrix",
		},
		{
			name:        "long-job-name",
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobs := cli.ReadJobsConfig(fmt.Sprintf("testdata/%s.yaml", tt.name))
			for _, branch := range jobs.Branches {
				output, err := cli.ConvertJobConfig(&jobs, branch)
				if tt.expectError {
					if err == nil {
						t.Fatalf("test %v expected an error, but did not receiev one", tt.name)
					}
					// there should be no generated file when an error occurs
					continue
				} else if err != nil {
					t.Fatalf("test %v did not expect an error, but received '%v'", tt.name, err)
				}
				testFile := fmt.Sprintf("testdata/%s.gen.yaml", tt.name)
				if os.Getenv("REFRESH_GOLDEN") == "true" {
					Write(output, testFile, bc.AutogenHeader)
				}
				if err := cli.CheckConfig(output, testFile, bc.AutogenHeader); err != nil {
					t.Fatal(err.Error())
				}
			}
		})
	}
}

func TestFilterReleaseBranchingJobs(t *testing.T) {
	testCases := []struct {
		name         string
		jobs         []Job
		filteredJobs []Job
	}{
		{
			name:         "filter an empty list of jobs",
			jobs:         []Job{},
			filteredJobs: []Job{},
		},
		{
			name: "filter enabled release branching jobs",
			jobs: []Job{
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
			filteredJobs: []Job{
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
			jobs: []Job{
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
			filteredJobs: []Job{},
		},
	}

	for _, tc := range testCases {
		expected := tc.filteredJobs
		actual := FilterReleaseBranchingJobs(tc.jobs)

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Filtered jobs do not	 match; actual: %v\n expected %v\n", actual, expected)
		}
	}
}

func TestMergeMaps(t *testing.T) {
	testCases := []struct {
		name     string
		mp1      map[string]string
		mp2      map[string]string
		expected map[string]string
	}{
		{
			name:     "combine two empty maps",
			mp1:      nil,
			mp2:      nil,
			expected: map[string]string{},
		},
		{
			name:     "the first map is empty",
			mp1:      map[string]string{},
			mp2:      map[string]string{"a": "aa", "b": "bb"},
			expected: map[string]string{"a": "aa", "b": "bb"},
		},
		{
			name:     "the second map is empty",
			mp1:      map[string]string{"a": "aa", "b": "bb"},
			mp2:      map[string]string{},
			expected: map[string]string{"a": "aa", "b": "bb"},
		},
		{
			name:     "two maps without duplicated keys",
			mp1:      map[string]string{"a": "aa"},
			mp2:      map[string]string{"b": "bb"},
			expected: map[string]string{"a": "aa", "b": "bb"},
		},
		{
			name:     "two maps with duplicated keys",
			mp1:      map[string]string{"a": "aa", "b": "b"},
			mp2:      map[string]string{"b": "bb"},
			expected: map[string]string{"a": "aa", "b": "bb"},
		},
	}

	for _, tc := range testCases {
		actual := mergeMaps(tc.mp1, tc.mp2)

		if !reflect.DeepEqual(tc.expected, actual) {
			t.Errorf("mergeMaps does not work as intended; actual: %v\n expected %v\n", actual, tc.expected)
		}
	}
}

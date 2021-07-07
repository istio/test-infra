/*
Copyright 2019 Istio Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"reflect"
	"testing"
)

func TestBranchJobSlices(t *testing.T) {
	testInstances := []struct {
		Name       string
		Values     []string
		BranchName string
		Out        []string
	}{
		{
			Name: "postsubmit",
			Values: []string{
				"foo_postsubmit",
			},
			BranchName: "test",
			Out: []string{
				"foo_test_postsubmit",
			},
		},
		{
			Name: "presubmit",
			Values: []string{
				"foo_presubmit",
			},
			BranchName: "test",
			Out: []string{
				"foo_test_presubmit",
			},
		},
	}

	for _, test := range testInstances {
		t.Run(test.Name, func(t *testing.T) {
			result := branchJobSlices(test.Values, test.BranchName)
			if !reflect.DeepEqual(result, test.Out) {
				t.Logf("Test \"%s\" failed: \n\t%+v \n\t\t not equal to \n\t%v", test.Name, result, test.Out)
				t.Fail()
			}
		})
	}
}

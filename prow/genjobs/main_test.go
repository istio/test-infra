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
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"istio.io/test-infra/prow/genjobs/cmd/genjobs"
)

const (
	testDir = "testdata"
)

func resolvePath(t *testing.T, filename string) string {
	name := filepath.Base(t.Name())
	return filepath.Join(testDir, strings.ToLower(name), filename)
}

func TestGenjobs(t *testing.T) {
	tests := []struct {
		name   string
		output string
		args   []string
		equal  bool
	}{
		{
			name:   "simple transform",
			output: "private.simple-transform.yaml",
			args:   []string{"--mapping=istio=istio-private"},
			equal:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			in := resolvePath(t, "")
			outE := resolvePath(t, test.output)

			expected, err := ioutil.ReadFile(outE)
			if err != nil {
				t.Errorf("Failed reading file %v: %v", outE, err)
			}

			tmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Errorf("Failed creating temp file: %v", err)
			}
			defer os.Remove(tmpDir)

			os.Args[0] = "genjobs"
			os.Args = append(os.Args, test.args...)
			os.Args = append(os.Args, "--input="+in, "--output="+tmpDir)
			genjobs.Main()

			outA := filepath.Join(tmpDir, test.output)

			actual, err := ioutil.ReadFile(outA)
			if err != nil {
				t.Errorf("Failed reading file %v: %v", outA, err)
			}

			equal := bytes.Equal(expected, actual)
			if equal != test.equal {
				t.Errorf("Expected output to be: %t.", test.equal)
			}
		})
	}
}

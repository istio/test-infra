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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var (
	tmpDir = "/tmp/test-infra_package-coverage"
)

func TestParseReport(t *testing.T) {
	exampleReport := "?   \tpilot/cmd\t[no test files]\nok  \tpilot/model\t1.3s\tcoverage: 90.2% of statements"
	reportFile := filepath.Join(tmpDir, "report")
	if err := ioutil.WriteFile(reportFile, []byte(exampleReport), 0644); err != nil {
		t.Errorf("Failed to write example report file, %v", err)
	}

	c := &codecovChecker{
		codeCoverage: make(map[string]float64),
		report:       reportFile,
	}

	if err := c.parseReport(); err != nil {
		t.Errorf("Failed to parse report, %v", err)
	} else {
		if len(c.codeCoverage) != 1 && c.codeCoverage["pilot/model"] != 90.2 {
			t.Error("Wrong result from parseReport()")
		}
	}
}

func TestSatisfiedRequirement(t *testing.T) {
	exampleRequirement := "pilot/model\t90"
	requirementFile := filepath.Join(tmpDir, "requirement1")
	if err := ioutil.WriteFile(requirementFile, []byte(exampleRequirement), 0644); err != nil {
		t.Errorf("Failed to write example requirement file, %v", err)
	}

	c := &codecovChecker{
		codeCoverage: map[string]float64{
			"pilot/model": 90.2,
		},
		requirement: requirementFile,
	}

	if err := c.checkRequirement(); err != nil {
		t.Errorf("Failed to check requirement, %v", err)
	} else {
		if len(c.failedPackage) != 0 {
			t.Error("Wrong result from checkRequirement()")
		}
	}
}

func TestMissRequirement(t *testing.T) {
	exampleRequirement := "pilot/model\t92.3"
	requirementFile := filepath.Join(tmpDir, "requirement1")
	if err := ioutil.WriteFile(requirementFile, []byte(exampleRequirement), 0644); err != nil {
		if err := ioutil.WriteFile(requirementFile, []byte(exampleRequirement), 0644); err != nil {
			t.Errorf("Failed to write example requirement file, %v", err)
		}
	}

	c := &codecovChecker{
		codeCoverage: map[string]float64{
			"pilot/model": 90.2,
		},
		requirement: requirementFile,
	}

	if err := c.checkRequirement(); err != nil {
		t.Errorf("Failed to check requirement, %v", err)
	} else {
		if len(c.failedPackage) != 1 {
			t.Error("Wrong result from checkRequirement()")
		}
	}
}

func TestMain(m *testing.M) {
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			fmt.Printf("Failed to remove tmpDir %s", tmpDir)
		}
	}()
	os.Exit(m.Run())
}

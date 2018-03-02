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
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const (
	tmpDirPrefix = "test-infra_package-coverage-"
)

var (
	tmpDir string
)

func TestParseReport(t *testing.T) {
	exampleReport := "?   \tpilot/cmd\t[no test files]\nok  \tpilot/model\t1.3s\tcoverage: 90.2% of statements"
	reportFile := filepath.Join(tmpDir, "report")
	if err := ioutil.WriteFile(reportFile, []byte(exampleReport), 0644); err != nil {
		t.Errorf("Failed to write example report file, %v", err)
	}

	c := &codecovChecker{
		codeCoverage:    make(map[string]float64),
		codeRequirement: make(map[string]float64),
		report:          reportFile,
	}

	if err := c.parseReport(); err != nil {
		t.Errorf("Failed to parse report, %v", err)
	} else {
		if len(c.codeCoverage) != 1 && c.codeCoverage["pilot/model"] != 90.2 {
			t.Error("Wrong result from parseReport()")
		}
	}
}

func TestParseRequirement(t *testing.T) {
	exampleRequirement := "Default:20\npilot/cmd:60\npilot/model:70 [75]"
	requirementFile := filepath.Join(tmpDir, "requirement")
	if err := ioutil.WriteFile(requirementFile, []byte(exampleRequirement), 0644); err != nil {
		t.Errorf("Failed to write example requirement file, %v", err)
	}

	c := &codecovChecker{
		codeCoverage:    make(map[string]float64),
		codeRequirement: make(map[string]float64),
		requirement:     requirementFile,
	}

	codeRequirementModel := map[string]float64{
		"Default":     20,
		"pilot/cmd":   60,
		"pilot/model": 70,
	}

	if err := c.parseRequirement(); err != nil {
		t.Errorf("Failed to parse requirement, %v", err)
	} else {
		if !reflect.DeepEqual(c.codeRequirement, codeRequirementModel) {
			t.Error("Wrong result from parseReport()")
		}
	}
}

func TestSatisfiedRequirement(t *testing.T) {
	exampleRequirement := "pilot/model:90"
	requirementFile := filepath.Join(tmpDir, "requirement2")
	if err := ioutil.WriteFile(requirementFile, []byte(exampleRequirement), 0644); err != nil {
		t.Errorf("Failed to write example requirement file, %v", err)
	}

	c := &codecovChecker{
		codeCoverage: map[string]float64{
			"pilot/model": 90.2,
		},
		codeRequirement: make(map[string]float64),
		requirement:     requirementFile,
	}

	if err := c.parseRequirement(); err != nil {
		t.Errorf("Failed to parse requirement, %v", err)
	}
	c.checkRequirement()
	if len(c.failedPackage) != 0 {
		t.Error("Wrong result from checkRequirement()")
	}

}

func TestMissRequirement(t *testing.T) {
	exampleRequirement := "pilot/model:92.3 [92.3]"
	requirementFile := filepath.Join(tmpDir, "requirement3")
	if err := ioutil.WriteFile(requirementFile, []byte(exampleRequirement), 0644); err != nil {
		if err := ioutil.WriteFile(requirementFile, []byte(exampleRequirement), 0644); err != nil {
			t.Errorf("Failed to write example requirement file, %v", err)
		}
	}

	c := &codecovChecker{
		codeCoverage: map[string]float64{
			"pilot/model": 90.2,
		},
		codeRequirement: make(map[string]float64),
		requirement:     requirementFile,
	}

	if err := c.parseRequirement(); err != nil {
		t.Errorf("Failed to parse requirement, %v", err)
	}
	c.checkRequirement()
	if len(c.failedPackage) != 1 {
		t.Error("Wrong result from checkRequirement()")
	}
}

func TestDefaultFailedCheck(t *testing.T) {
	c := &codecovChecker{
		codeCoverage: map[string]float64{
			"pilot/model": 15,
		},
		codeRequirement: map[string]float64{
			"Default": 20,
		},
	}

	c.checkRequirement()
	if len(c.failedPackage) != 1 {
		t.Error("Wrong result from checkRequirement()")
	}
}

func TestPassCheck(t *testing.T) {
	exampleReport := "ok  \tpilot/model\t1.3s\tcoverage: 90.2% of statements"
	reportFile := filepath.Join(tmpDir, "report4")
	if err := ioutil.WriteFile(reportFile, []byte(exampleReport), 0644); err != nil {
		t.Errorf("Failed to write example report file, %v", err)
	}

	exampleRequirement := "pilot/model:89"
	requirementFile := filepath.Join(tmpDir, "requirement4")
	if err := ioutil.WriteFile(requirementFile, []byte(exampleRequirement), 0644); err != nil {
		if err := ioutil.WriteFile(requirementFile, []byte(exampleRequirement), 0644); err != nil {
			t.Errorf("Failed to write example requirement file, %v", err)
		}
	}

	c := &codecovChecker{
		codeCoverage:    make(map[string]float64),
		codeRequirement: make(map[string]float64),
		report:          reportFile,
		requirement:     requirementFile,
		bucket:          "fake",
	}

	// No other error code, code only show gcs upload failed which is expected
	if code := c.checkPackageCoverage(); code != 3 {
		t.Errorf("Unexpected return code, expected: %d, actual: %d", 3, code)
	}
}

func TestFailedCheck(t *testing.T) {
	exampleReport := "?   \tpilot/cmd\t[no test files]\nok  \tpilot/model\t1.3s\tcoverage: 90.2% of statements"
	reportFile := filepath.Join(tmpDir, "report5")
	if err := ioutil.WriteFile(reportFile, []byte(exampleReport), 0644); err != nil {
		t.Errorf("Failed to write example report file, %v", err)
	}

	exampleRequirement := "pilot/model:93 [93]"
	requirementFile := filepath.Join(tmpDir, "requirement5")
	if err := ioutil.WriteFile(requirementFile, []byte(exampleRequirement), 0644); err != nil {
		if err := ioutil.WriteFile(requirementFile, []byte(exampleRequirement), 0644); err != nil {
			t.Errorf("Failed to write example requirement file, %v", err)
		}
	}

	c := &codecovChecker{
		codeCoverage:    make(map[string]float64),
		codeRequirement: make(map[string]float64),
		report:          reportFile,
		requirement:     requirementFile,
	}

	if code := c.checkPackageCoverage(); code != 2 {
		t.Errorf("Unexpected return code, expected: %d, actual: %d", 2, code)
	}
}

func TestWriteRequirementFromReport(t *testing.T) {
	exampleReport := "?   \tpilot/cmd\t[no test files]\nok  \tpilot/model\t1.3s\tcoverage: 90.2% of statements"
	reportFile := filepath.Join(tmpDir, "report5")
	if err := ioutil.WriteFile(reportFile, []byte(exampleReport), 0644); err != nil {
		t.Errorf("Failed to write example report file, %v", err)
	}

	expectedRequirement := fmt.Sprintf(
		"Default:%d [%.1f]\npilot/cmd:0 [0.0]\npilot/model:90 [90.2]\n", int(defaultThreshold), defaultThreshold)

	requirementFile := filepath.Join(tmpDir, "requirement6")

	c := &codecovChecker{
		codeCoverage:    make(map[string]float64),
		codeRequirement: make(map[string]float64),
		report:          reportFile,
		requirement:     requirementFile,
	}

	if code := c.writeRequirementFromReport(); code != 0 {
		t.Errorf("Unexpected return code, expected: %d, actual: %d", 0, code)
	}
	data, err := ioutil.ReadFile(requirementFile)
	if err != nil {
		t.Errorf("unable to read requirement file, %v", err)
	}
	if string(data) != expectedRequirement {
		t.Errorf("\n%s\nshould match expected\n%s", string(data), expectedRequirement)
	}
}

func TestMain(m *testing.M) {
	var err error
	if tmpDir, err = ioutil.TempDir("", tmpDirPrefix); err != nil {
		log.Printf("Failed to create tmp directory: %s, %s", tmpDir, err)
		os.Exit(4)
	}

	exitCode := m.Run()

	if err := os.RemoveAll(tmpDir); err != nil {
		log.Printf("Failed to remove tmpDir %s", tmpDir)
	}

	os.Exit(exitCode)
}

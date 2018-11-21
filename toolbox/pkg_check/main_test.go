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
	reportFile := filepath.Join(tmpDir, "reportFile")
	if err := ioutil.WriteFile(reportFile, []byte(exampleReport), 0644); err != nil {
		t.Errorf("Failed to write example reportFile file, %v", err)
	}

	codeCoverage, err := parseReport(reportFile, false)
	if err != nil {
		t.Errorf("Failed to parse reportFile, %v", err)
	} else {
		if len(codeCoverage) != 1 && codeCoverage["pilot/model"] != 90.2 {
			t.Error("Wrong result from parseReport()")
		}
	}
}

func TestParseHtml(t *testing.T) {
	exampleHtml :=
		"<html>\n" +
			"    <option value=\"file2\">istio.io/istio/galley/pkg/crd/validation/endpoint.go (62.8%)</option>\n" +
			"\n" +
			"    <option value=\"file3\">istio.io/istio/galley/pkg/crd/validation/monitoring.go (61.2%)</option>\n" +
			"</html>\n"
	reportFile := filepath.Join(tmpDir, "reportFile")
	if err := ioutil.WriteFile(reportFile, []byte(exampleHtml), 0644); err != nil {
		t.Errorf("Failed to write example reportFile file, %v", err)
	}

	codeCoverage, err := parseReport(reportFile, true)
	if err != nil {
		t.Errorf("Failed to parse reportFile, %v", err)
	} else {
		if len(codeCoverage) != 2 {
			t.Error("Wrong result count from parseReport()")
		}
		if codeCoverage["istio.io/istio/galley/pkg/crd/validation/endpoint.go"] != 62.8 {
			t.Error("Wrong result from parseReport()")
		}
		if codeCoverage["istio.io/istio/galley/pkg/crd/validation/monitoring.go"] != 61.2 {
			t.Error("Wrong result from parseReport()")
		}
	}
}

func TestFindDelta(t *testing.T) {
	dettas := findDelta(
		// report
		map[string]float64{
			"P1": 30,
			"P2": 90,
			"P3": 100,
			"P4": 90,
		},
		// baseline
		map[string]float64{
			"P1": 50,
			"P2": 60,
			"P3": 100,
			"P5": 60,
		},
	)
	expected := map[string]float64{
		"P1": -20,
		"P2": 30,
		"P3": 0,
		"P4": 90,
		"P5": -60,
	}
	if !reflect.DeepEqual(dettas, expected) {
		t.Errorf("Actual: %s; expected: %s", fmt.Sprint(dettas), fmt.Sprint(expected))
	}
}

func TestCheckDeltaError(t *testing.T) {
	code := checkDelta(
		// Delta
		map[string]float64{
			"P1": -20,
			"P2": 30,
			"P3": 0,
			"P4": 90,
			"P5": -60,
		},
		// report
		map[string]float64{
			"P1": 30,
			"P2": 90,
			"P3": 100,
			"P4": 90,
		},
		// baseline
		map[string]float64{
			"P1": 50,
			"P2": 60,
			"P3": 100,
			"P5": 60,
		})
	if code != ThresholdExceeded {
		t.Errorf("Expecting return code 2, got %d", code)
	}
}

func TestCheckDeltaGood(t *testing.T) {
	code := checkDelta(
		// Delta
		map[string]float64{
			"P1": -1,
			"P2": 30,
			"P3": 0,
			"P4": 90,
		},
		// report
		map[string]float64{
			"P1": 30,
			"P2": 90,
			"P3": 100,
			"P4": 90,
		},
		// baseline
		map[string]float64{
			"P1": 31,
			"P2": 60,
			"P3": 100,
		})
	if code != NoError {
		t.Errorf("Expecting return code 0, got %d", code)
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

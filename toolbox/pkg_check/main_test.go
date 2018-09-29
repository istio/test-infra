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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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

	codeCoverage, err := parseReport(reportFile)
	if err != nil {
		t.Errorf("Failed to parse reportFile, %v", err)
	} else {
		if len(codeCoverage) != 1 && codeCoverage["pilot/model"] != 90.2 {
			t.Error("Wrong result from parseReport()")
		}
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

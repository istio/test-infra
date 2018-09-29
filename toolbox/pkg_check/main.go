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
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

const (
	// Using 2% less than reportFile for requirement
	thresholdDelta = 2
)

var (
	reportFile   = flag.String("report_file", "codecov.reportFile", "Package code coverage reportFile.")
	baselineFile = flag.String("baseline_file", "", "Package code coverage baseline.")
)

//Report example: "ok   istio.io/mixer/adapter/denyChecker      0.023s  coverage: 100.0% of statements"
//expected output: c.codeCoverage["istio.io/mixer/adapter/denyChecker"] = 100
//Report example: "?    istio.io/mixer/adapter/denyChecker/config       [no test files]"
//Report example: c.codeCoverage["istio.io/mixer/adapter/denyChecker/config"] = 0
func parseReportLine(line string) (string, float64, error) {
	regOK := regexp.MustCompile(`(ok  )\t(.*)\t(.*)\tcoverage: (.*) of statements`)
	regNoTest := regexp.MustCompile(`(\?   )\t(.*)\t\[no test files\]`)
	pkgPos := 2
	numPos := 4

	if m := regOK.FindStringSubmatch(line); len(m) != 0 {
		n, err := strconv.ParseFloat(strings.TrimSuffix(m[numPos], "%"), 64)
		if err != nil {
			return "", 0, err
		}
		return m[pkgPos], n, nil
	} else if m := regNoTest.FindStringSubmatch(line); len(m) != 0 {
		return m[pkgPos], 0, nil
	}
	return "", 0, fmt.Errorf("unclear line from reportFile: %s", line)
}

func parseReport(filename string) (map[string]float64, error) {
	coverage := make(map[string]float64)

	f, err := os.Open(filename)
	if err != nil {
		glog.Errorf("Failed to open file %s, %v", filename, err)
		return coverage, err
	}
	defer func() {
		if err = f.Close(); err != nil {
			glog.Warningf("Failed to close file %s, %v", filename, err)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if pkg, cov, err := parseReportLine(scanner.Text()); err == nil {
			coverage[pkg] = cov
		}
	}
	return coverage, scanner.Err()
}

func findDelta(codeCoverage, baseline map[string]float64) {
	good := make(map[string]float64)
	bad := make(map[string]float64)

	for pkg, cov := range codeCoverage {
		var delta float64
		if base, exist := baseline[pkg]; !exist {
			//There is no entry in the baseline.
			delta = cov
		} else {
			delete(baseline, pkg)
			delta = cov - base
		}
		if delta+thresholdDelta < 0 {
			bad[pkg] = delta
		} else {
			good[pkg] = delta
		}
	}
	// Find the remining packages that exist in baseline but not in report.
	for pkg, base := range baseline {
		delete(baseline, pkg)
		bad[pkg] = 0 - base
	}

	for pkg, delta := range good {
		glog.Infof("Package %s changed: %f", pkg, delta)
	}
	for pkg, delta := range bad {
		glog.Errorf("Package %s dropped %f", pkg, delta)
	}
}

func checkBaseline(reportFile, baselineFile string) (code int) {
	codeCoverage, err := parseReport(reportFile)
	if err != nil {
		glog.Error(err)
		return 1 //Error code 1: Parse file failure
	}
	baseline, err := parseReport(baselineFile)
	if err != nil {
		glog.Error(err)
		return 1 //Error code 1: Parse file failure
	}
	findDelta(codeCoverage, baseline)
	return 0
}

func main() {
	flag.Parse()
	os.Exit(checkBaseline(*reportFile, *baselineFile))
}

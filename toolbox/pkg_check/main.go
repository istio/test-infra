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

var (
	reportFile   = flag.String("report_file", "codecov.reportFile", "Package code coverage reportFile.")
	baselineFile = flag.String("baseline_file", "", "Package code coverage baseline.")
	threshold    = flag.Float64("threshold", 5, "Coverage drop threshold. Trigger error if any package drops more than this.")
	html         = flag.Bool("html", false, "Whether the report files are in html")
)

const (
	// NoError is return code for no error
	NoError = 0
	// ThresholdExceeded is return code in case codecov threshold is exceeded.
	ThresholdExceeded = 2
)

func parseReportLine(line string, html bool) (string, float64, error) {
	// <option value="file0">istio.io/istio/galley/cmd/shared/shared.go (0.0%)</option>
	reg := regexp.MustCompile(` *<option value=\"(.*)\">(.*) \((.*)%\)</option>`)
	if html {
		if m := reg.FindStringSubmatch(line); len(m) != 0 {
			cov, err := strconv.ParseFloat(m[3], 64)
			if err != nil {
				return "", 0, err
			}
			return m[2], cov, nil
		}
		return "", 0, fmt.Errorf("no coverage in %s", line)

	}
	// TODO Remove when we switch to all HTML reporting
	//Report example: "ok   istio.io/mixer/adapter/denyChecker      0.023s  coverage: 100.0% of statements"
	//expected output: c.codeCoverage["istio.io/mixer/adapter/denyChecker"] = 100
	//Report example: "?    istio.io/mixer/adapter/denyChecker/config       [no test files]"
	//Report example: c.codeCoverage["istio.io/mixer/adapter/denyChecker/config"] = 0
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

func parseReport(filename string, html bool) (map[string]float64, error) {
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
		if pkg, cov, err := parseReportLine(scanner.Text(), html); err == nil {
			coverage[pkg] = cov
		}
	}
	return coverage, scanner.Err()
}

func findDelta(report, baseline map[string]float64) map[string]float64 {
	deltas := make(map[string]float64)

	for pkg, cov := range report {
		deltas[pkg] = cov - baseline[pkg]
	}
	// Find the remaining packages that exist in baseline but not in report.
	for pkg, base := range baseline {
		if _, exist := report[pkg]; !exist {
			deltas[pkg] = 0 - base
		}
	}
	return deltas
}

func checkDelta(deltas, report, baseline map[string]float64) int {
	code := NoError

	// First print all coverage change.
	for pkg, delta := range deltas {
		glog.Infof("Coverage change: %s:%f%% (%f%% to %f%%)", pkg, delta, baseline[pkg], report[pkg])
	}

	// Then generate errors for reduced coverage.
	for pkg, delta := range deltas {
		if delta+*threshold < 0 {
			glog.Errorf("Coverage dropped: %s:%f%% (%f%% to %f%%)", pkg, delta, baseline[pkg], report[pkg])
			code = ThresholdExceeded
		}
	}
	return code
}

func checkBaseline(reportFile, baselineFile string) int {
	report, err := parseReport(reportFile, *html)
	if err != nil {
		glog.Error(err)
		return 1 //Error code 1: Parse file failure
	}
	baseline, err := parseReport(baselineFile, *html)
	if err != nil {
		glog.Error(err)
		return 1 //Error code 1: Parse file failure
	}
	deltas := findDelta(report, baseline)
	return checkDelta(deltas, report, baseline)
}

func main() {
	flag.Parse()
	os.Exit(checkBaseline(*reportFile, *baselineFile))
}

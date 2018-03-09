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
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

const (
	defaultValue     = "Default"
	defaultThreshold = 80.0
	// Using 2% less than report for requirement
	thresholdDelta = 2
	googleACEnv    = "GOOGLE_APPLICATION_CREDENTIALS"
)

var (
	reportFile           = flag.String("report_file", "codecov.report", "Package code coverage report.")
	requirementFile      = flag.String("requirement_file", "codecov.requirement", "Package code coverage requirement.")
	gcsBucket            = flag.String("bucket", "istio-code-coverage", "gcs bucket")
	writeRequirement     = flag.Bool("write_requirement", false, "Write requirement file from report")
	defaultThresholdFlag = flag.Float64("default_threshold", defaultThreshold, "Default threshold for new packages")
	jobIdentifier        = flag.String("job_name", "", "Name of job to store data")
	buildID              = flag.String("build_id", "", "Build ID")
	serviceAccountJSON   = flag.String("service_account", "", "Path to the service account key")
)

type uploader interface {
	upload(ctx context.Context, dest, data string) error
}

type codecovChecker struct {
	codeCoverage    map[string]float64
	codeRequirement map[string]float64
	report          string
	requirement     string
	failedPackage   []string
	defautThreshold float64
	thresholdDelta  float64
	// Used for uploading data to a bucket for metrics
	buildID       string
	jobIdentifier string
	storage       uploader
}

type googleStorageUploader struct {
	bucket             string
	serviceAccountJSON string
}

func (g *googleStorageUploader) upload(ctx context.Context, dest, data string) error {
	if g.serviceAccountJSON != "" {
		if err := os.Setenv(googleACEnv, g.serviceAccountJSON); err != nil {
			return err
		}
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		glog.Warningf("Failed to get storage client")
		return err
	}

	w := client.Bucket(g.bucket).Object(dest).NewWriter(ctx)
	defer func() {
		if err = w.Close(); err != nil {
			glog.Errorf("Failed to close gcs writer file, %v", err)
		}
	}()
	if _, err = w.Write([]byte(data)); err != nil {
		glog.Errorf("Failed to write coverage to gcs")
		return err
	}
	glog.Infof("Successfully uploaded codecov.report %s", dest)
	return nil
}

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
			glog.Errorf("Failed to parse coverage to float64 for package %s: %s, %v",
				m[pkgPos], m[numPos], err)
			return "", 0, err
		}
		return m[pkgPos], n, nil
	} else if m := regNoTest.FindStringSubmatch(line); len(m) != 0 {
		return m[pkgPos], 0, nil
	}
	return "", 0, fmt.Errorf("unclear line from report: %s", line)
}

func (c *codecovChecker) parseReport() error {
	f, err := os.Open(c.report)
	if err != nil {
		glog.Errorf("Failed to open report file %s, %v", c.report, err)
		return err
	}
	defer func() {
		if err = f.Close(); err != nil {
			glog.Warningf("Failed to close file %s, %v", c.report, err)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if pkg, cov, err := parseReportLine(scanner.Text()); err != nil {
			glog.Warningf("Failed to parse this line from report file: %s, %v", scanner.Text(), err)
		} else {
			c.codeCoverage[pkg] = cov
		}
	}
	return scanner.Err()
}

//Requirement example: "istio.io/mixer/adapter/denyChecker:99 [100]"
//Expected output: parts = {"istio.io/mixer/adapter/denyChecker", "99"}
//Default requirement example: "Default:20"
//Expected output: c.codeRequirement["Default"] = 20
func parseRequirementLine(line string) (string, float64, error) {
	reg := regexp.MustCompile(`(.*):([0-9]{1,2}|100)( \[([0-9]{1,2}|100)\])?`)
	lenWithCov := 5
	lenWithoutCov := 3
	pkgPos := 1
	reqPos := 2

	m := reg.FindStringSubmatch(line)
	if len(m) == lenWithCov || len(m) == lenWithoutCov {
		n, err := strconv.ParseFloat(m[reqPos], 64)
		if err != nil {
			return "", math.MaxFloat64, fmt.Errorf("failed to parse requirement to float64 for package %s: %s, %v",
				m[pkgPos], m[reqPos], err)
		}
		return m[pkgPos], n, nil
	}
	return "", math.MaxFloat64, fmt.Errorf("unclear line from requirement: %s", line)
}

func (c *codecovChecker) parseRequirement() error {
	f, err := os.Open(c.requirement)
	if err != nil {
		glog.Errorf("Failed to open requirement file, %s, %v", c.requirement, err)
		return err
	}
	defer func() {
		if err = f.Close(); err != nil {
			glog.Errorf("Failed to close file %s, %v", c.requirement, err)
		}
	}()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		pkg, req, err := parseRequirementLine(scanner.Text())
		if err != nil {
			glog.Errorf("Failed to parse this line from requirement file: %s, %v", scanner.Text(), err)
		}
		c.codeRequirement[pkg] = req
	}
	return scanner.Err()
}

func (c *codecovChecker) checkRequirement() {
	for pkg, cov := range c.codeCoverage {
		if req, exist := c.codeRequirement[pkg]; !exist {
			//There is no entry for this package in requirement file, set default requirement
			if defaultReq, exist := c.codeRequirement[defaultValue]; !exist {
				c.failedPackage = append(c.failedPackage, fmt.Sprintf("%s\t%.2f\tNo pkg or default requirement", pkg, cov))
			} else {
				if cov < defaultReq {
					c.failedPackage = append(c.failedPackage, fmt.Sprintf("%s\t%.2f\t%.2f(default)", pkg, cov, c.codeRequirement["Default"]))
				}
			}
		} else {
			if cov < req {
				c.failedPackage = append(c.failedPackage, fmt.Sprintf("%s\t%.2f\t%.2f", pkg, cov, req))
			}
		}
	}
}

func (c *codecovChecker) uploadCoverage() error {
	coverageString := ""
	for p, c := range c.codeCoverage {
		coverageString += fmt.Sprintf("%s\t%.2f\n", p, c)
	}

	dest := fmt.Sprintf("%s/%s", c.jobIdentifier, c.buildID)
	if c.buildID == "" || c.jobIdentifier == "" {
		glog.Errorf("Missing build info: BUILD_ID: \"%s\", JOB_NAME: \"%s\"\n", c.buildID, c.jobIdentifier)
		return errors.New("missing build info")
	}
	return c.storage.upload(context.Background(), dest, coverageString)
}

func (c *codecovChecker) writeRequirementFromReport() (code int) {
	if err := c.parseReport(); err != nil {
		glog.Errorf("Failed to parse report, %v", err)
		return 1 //Error code 1: Parse file failure
	}

	var sortedPkgs []string
	for k := range c.codeCoverage {
		sortedPkgs = append(sortedPkgs, k)
	}

	sort.Strings(sortedPkgs)

	f, err := os.Create(c.requirement)
	if err != nil {
		glog.Errorf("unable to create file %s", c.requirement)
		return 4 //Error code 4: Unable to create requirement file
	}
	defer func() {
		if err := f.Close(); err != nil {
			glog.Warningf("unable to close requirement file")
		}
	}()

	w := bufio.NewWriter(f)
	// Writing header
	header := fmt.Sprintf("%s:%d [%.1f]\n", defaultValue, int(defaultThreshold), defaultThreshold)
	if _, err := w.WriteString(header); err != nil {
		glog.Errorf("unable to write requirement file")
		return 5 //Error code 5: unable to write to requirement file
	}
	for _, pkg := range sortedPkgs {
		percent := c.codeCoverage[pkg]
		threshold := int(math.Max(0, percent-c.thresholdDelta))
		if percent == 100.0 {
			threshold = 100
		}
		if _, err := w.WriteString(fmt.Sprintf("%s:%d [%.1f]\n", pkg, threshold, percent)); err != nil {
			glog.Errorf("unable to write requirement file")
			return 5 //Error code 5: unable to write to requirement file
		}
	}

	if err := w.Flush(); err != nil {
		glog.Warningf("unable to flush requirement file")
	}
	return 0
}

func (c *codecovChecker) checkPackageCoverage() (code int) {

	if err := c.parseReport(); err != nil {
		glog.Errorf("Failed to parse report, %v", err)
		return 1 //Error code 1: Parse file failure
	}

	if err := c.parseRequirement(); err != nil {
		glog.Errorf("Failed to parse requirement, %v", err)
		return 1 //Error code 1: Parse file failure
	}

	c.checkRequirement()

	if c.storage != nil {
		if err := c.uploadCoverage(); err != nil {
			// Failing silently
			glog.Warningf("Failed to upload coverage, %v", err)
		}
	}

	if len(c.failedPackage) == 0 {
		glog.Infof("All packages passed code coverage requirements!")
	} else {
		glog.Errorf("Following package(s) failed to meet requirements:\n\tPackage Name\t\tActual Coverage\t\tRequirement\n")
		for _, p := range c.failedPackage {
			glog.Errorf(p)
		}
		return 2 //Error code 2: Unsatisfied coverage requirement
	}
	return 0
}

func main() {
	flag.Parse()

	var s uploader
	if *gcsBucket != "" {
		s = &googleStorageUploader{
			bucket:             *gcsBucket,
			serviceAccountJSON: *serviceAccountJSON,
		}
	}

	c := &codecovChecker{
		codeCoverage:    make(map[string]float64),
		codeRequirement: make(map[string]float64),
		report:          *reportFile,
		requirement:     *requirementFile,
		defautThreshold: *defaultThresholdFlag,
		thresholdDelta:  thresholdDelta,
		buildID:         *buildID,
		jobIdentifier:   *jobIdentifier,
		storage:         s,
	}

	if *writeRequirement {
		os.Exit(c.writeRequirementFromReport())
	}
	os.Exit(c.checkPackageCoverage())
}

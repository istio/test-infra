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
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"sort"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

const (
	jobName          = "JOB_NAME"
	buildID          = "BUILD_ID"
	defaultValue     = "Default"
	defaultThreshold = 80.0
)

var (
	reportFile           = flag.String("report_file", "codecov.report", "Package code coverage report.")
	requirementFile      = flag.String("requirement_file", "codecov.requirement", "Package code coverage requirement.")
	gcsBucket            = flag.String("bucket", "istio-code-coverage", "gcs bucket")
	writeRequirement     = flag.Bool("write_requirement", false, "Write requirement file from report")
	defaultThresholdFlag = flag.Float64("default_threshold", defaultThreshold, "Default threshold for new packages")
)

type codecovChecker struct {
	codeCoverage    map[string]float64
	codeRequirement map[string]float64
	report          string
	requirement     string
	failedPackage   []string
	bucket          string
	defautThreshold float64
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
			log.Printf("Failed to parse coverage to float64 for package %s: %s, %v",
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
		log.Printf("Failed to open report file %s, %v", c.report, err)
		return err
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Printf("Failed to close file %s, %v", c.report, err)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if pkg, cov, err := parseReportLine(scanner.Text()); err != nil {
			log.Printf("Failed to parse this line from report file: %s, %v", scanner.Text(), err)
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
		log.Printf("Failed to open requirement file, %s, %v", c.requirement, err)
		return err
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Printf("Failed to close file %s, %v", c.requirement, err)
		}
	}()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		pkg, req, err := parseRequirementLine(scanner.Text())
		if err != nil {
			log.Printf("Failed to parse this line from requirement file: %s, %v", scanner.Text(), err)
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
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Print("Failed to get storage client")
		return err
	}

	jobName := os.Getenv(jobName)
	buildID := os.Getenv(buildID)
	if buildID == "" || jobName == "" {
		log.Printf("Missing build info: BUILD_ID: \"%s\", JOB_NAME: \"%s\"\n", buildID, jobName)
		return errors.New("missing build info")
	}

	object := jobName + "/" + buildID

	coverageString := ""
	for p, c := range c.codeCoverage {
		coverageString += fmt.Sprintf("%s\t%.2f\n", p, c)
	}

	w := client.Bucket(c.bucket).Object(object).NewWriter(ctx)
	if _, err = w.Write([]byte(coverageString)); err != nil {
		log.Print("Failed to write coverage to gcs")
		return err
	}

	defer func() {
		if err = w.Close(); err != nil {
			log.Printf("Failed to close gcs writer file, %v", err)
		}
	}()

	log.Printf("Successfully upload codecov.report %s", object)
	return nil
}

func (c *codecovChecker) writeRequirementFromReport() (code int) {

	if err := c.parseReport(); err != nil {
		log.Printf("Failed to parse report, %v", err)
		return 1 //Error code 1: Parse file failure
	}

	var sortedPkgs []string
	for k := range c.codeCoverage {
		sortedPkgs = append(sortedPkgs, k)
	}

	sort.Strings(sortedPkgs)

	f, err := os.Create(c.requirement)
	if err != nil {
		log.Printf("unable to create file %s", c.requirement)
		return 4 //Error code 4: Unable to create requirement file
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	// Writing default
	w.WriteString(fmt.Sprintf("%s:%d [%.1f]\n", defaultValue, int(defaultThreshold), defaultThreshold))
	for _, pkg := range sortedPkgs {
		percent := c.codeCoverage[pkg]
		if _, err := w.WriteString(fmt.Sprintf("%s:%d [%.1f]\n", pkg, int(percent), percent)); err != nil {
			log.Printf("unable to print ")
			return 5 //Error code 5: unable to write to requirement file
		}

	}

	w.Flush()
	return 0
}

func (c *codecovChecker) checkPackageCoverage() (code int) {
	if c.bucket != "" {
		defer func() {
			if err := c.uploadCoverage(); err != nil {
				log.Printf("Failed to upload coverage, %v", err)
				if code == 0 {
					code = 3 //If no other error code, Error code 3: Failed to upload coverage
				}
			}

		}()
	}

	if err := c.parseReport(); err != nil {
		log.Printf("Failed to parse report, %v", err)
		return 1 //Error code 1: Parse file failure
	}

	if err := c.parseRequirement(); err != nil {
		log.Printf("Failed to parse requirement, %v", err)
		return 1 //Error code 1: Parse file failure
	}

	c.checkRequirement()

	if len(c.failedPackage) == 0 {
		log.Println("All packages passed code coverage requirements!")
	} else {
		log.Printf("Following package(s) failed to meet requirements:\n\tPackage Name\t\tActual Coverage\t\tRequirement\n")
		for _, p := range c.failedPackage {
			log.Println(p)
		}
		return 2 //Error code 2: Unsatisfied coverage requirement
	}
	return 0
}

func main() {
	flag.Parse()

	c := &codecovChecker{
		codeCoverage:    make(map[string]float64),
		codeRequirement: make(map[string]float64),
		report:          *reportFile,
		requirement:     *requirementFile,
		bucket:          *gcsBucket,
		defautThreshold: *defaultThresholdFlag,
	}

	if *writeRequirement {
		os.Exit(c.writeRequirementFromReport())
	}
	os.Exit(c.checkPackageCoverage())
}

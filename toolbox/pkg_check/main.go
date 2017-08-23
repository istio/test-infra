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
	"os"
	"regexp"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

const (
	jobName = "JOB_NAME"
	buildID = "BUILD_ID"
)

var (
	reportFile      = flag.String("report_file", "codecov.report", "Package code coverage report.")
	requirementFile = flag.String("requirement_file", "codecov.requirement", "Package code coverage requirement.")
	gcsBucket       = flag.String("bucket", "istio-code-coverage", "gcs bucket")
)

type codecovChecker struct {
	codeCoverage    map[string]float64
	codeRequirement map[string]float64
	report          string
	requirement     string
	failedPackage   []string
	bucket          string
}

func (c *codecovChecker) parseReport() error {
	f, err := os.Open(c.report)
	if err != nil {
		log.Printf("Failed to open report file %s", c.report)
		return err
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Printf("Failed to close file %s, %v", c.report, err)
		}
	}()

	scanner := bufio.NewScanner(f)

	//Report example: "ok   istio.io/mixer/adapter/denyChecker      0.023s  coverage: 100.0% of statements"
	//expected output: c.codeCoverage["istio.io/mixer/adapter/denyChecker"] = 100
	//Report example: "?    istio.io/mixer/adapter/denyChecker/config       [no test files]"
	//Report example: c.codeCoverage["istio.io/mixer/adapter/denyChecker/config"] = 0
	regOK := regexp.MustCompile(`(ok  )\t(.*)\t(.*)\tcoverage: (.*) of statements`)
	regNoTest := regexp.MustCompile(`(\?   )\t(.*)\t\[no test files\]`)
	for scanner.Scan() {
		if m := regOK.FindStringSubmatch(scanner.Text()); len(m) != 0 {

			if n, err := strconv.ParseFloat(strings.TrimSuffix(m[4], "%"), 64); err != nil {
				log.Printf("Failed to parse coverage to float64 for package %s: %s", m[2], m[4])
				return err
			} else {
				c.codeCoverage[m[2]] = n
			}
		} else if m := regNoTest.FindStringSubmatch(scanner.Text()); len(m) != 0 {
			c.codeCoverage[m[2]] = 0
		} else {
			log.Printf("Unclear line from report: %s", scanner.Text())
		}
	}
	return scanner.Err()
}

func (c *codecovChecker) parseRequirement() error {
	f, err := os.Open(c.requirement)
	if err != nil {
		log.Printf("Failed to open requirement file, %s", c.requirement)
		return err
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Printf("Failed to close file %s, %v", c.requirement, err)
		}
	}()

	scanner := bufio.NewScanner(f)

	//Requirement example: "istio.io/mixer/adapter/denyChecker:99 [100]"
	//Expected output: parts = {"istio.io/mixer/adapter/denyChecker", "99"}
	//Default requirement example: "Default:20"
	//Expected output: c.codeRequirement["Default"] = 20
	reg := regexp.MustCompile(`(.*):([0-9]{1,2}|100)( \[([0-9]{1,2}|100)\])?`)
	for scanner.Scan() {
		m := reg.FindStringSubmatch(scanner.Text())
		if len(m) == 5 || len(m) == 3 {
			if n, err := strconv.ParseFloat(m[2], 64); err != nil {
				log.Printf("Failed to parse requirement to float64 for package %s: %s", m[1], m[2])
				continue
			} else {
				c.codeRequirement[m[1]] = n
			}
		} else {
			log.Printf("Unclear line from requirement: %s", scanner.Text())
		}
	}

	return scanner.Err()
}

func (c *codecovChecker) checkRequirement() {
	for pkg, cov := range c.codeCoverage {
		if req, exist := c.codeRequirement[pkg]; !exist {
			//There is no entry for this package in requirement file, set default requirement
			if cov < c.codeRequirement["Default"] {
				c.failedPackage = append(c.failedPackage, fmt.Sprintf("%s\t%.2f\t%.2f(default)", pkg, cov, c.codeRequirement["Default"]))
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

func (c *codecovChecker) checkPackageCoverage() (code int) {
	defer func() {
		if err := c.uploadCoverage(); err != nil {
			log.Printf("Failed to upload coverage, %v", err)
			if code == 0 {
				code = 3 //If no other error code, Error code 3: Failed to upload coverage
			}
		}
	}()

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
		log.Printf("Following package(s) failed to meet requirements:\nPackage Name\t\tActual Coverage\t\tRequirement\n")
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
	}

	os.Exit(c.checkPackageCoverage())
}

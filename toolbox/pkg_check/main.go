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
	codeCoverage  map[string]float64
	report        string
	requirement   string
	failedPackage []string
	bucket        string
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

	//Report example: "ok  	istio.io/mixer/adapter/denyChecker	0.023s	coverage: 100.0% of statements"
	//Report example: "?   istio.io/mixer/adapter/denyChecker/config	[no test files]"
	//expected output: c.codeCoverage["istio.io/mixer/adapter/denyChecker"] = 100
	regOK := regexp.MustCompile(`(ok  )\t(.*)\t(.*)\tcoverage: (.*) of statements`)
	regQ := regexp.MustCompile(`(\\?   )\t(.*)\t(.*)`)
	for scanner.Scan() {
		fmt.Println(scanner.Text()) //Print report back to stdout
		if m := regOK.FindStringSubmatch(scanner.Text()); len(m) != 0 {
			n, err := strconv.ParseFloat(strings.TrimSuffix(m[4], "%"), 64)
			if err != nil {
				log.Printf("Failed to parse coverage to float64 for package %s: %s", m[2], m[4])
				return err
			}
			c.codeCoverage[m[2]] = n
		} else if m := regQ.FindStringSubmatch(scanner.Text()); len(m) != 0 {
			log.Printf("No test file in this package: %s", m[2])
		} else {
			log.Printf("Unclear line from report: %s", scanner.Text())
		}
	}
	return scanner.Err()
}

func (c *codecovChecker) checkRequirement() error {
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

	//Requirement example: "istio.io/mixer/adapter/denyChecker	99"
	//Expected output: parts = {"istio.io/mixer/adapter/denyChecker", "99"}
	reg := regexp.MustCompile(`(.*)\t(.*)`)
	for scanner.Scan() {
		m := reg.FindStringSubmatch(scanner.Text())
		if len(m) != 0 {
			if cov, exist := c.codeCoverage[m[1]]; exist {
				re, err := strconv.ParseFloat(m[2], 64)
				if err != nil {
					log.Printf("Failed to get requirement for package %s: %s", m[1], m[2])
					return err
				}
				if cov < re {
					c.failedPackage = append(c.failedPackage, fmt.Sprintf("%s\t%.2f\t%s", m[1], cov, m[2]))
				}

			} else {
				c.failedPackage = append(c.failedPackage, fmt.Sprintf("%s\t%s\t%s", m[1], "0.0", m[2]))
			}
		} else {
			log.Printf("Unclear line from requirement: %s", scanner.Text())
		}
	}

	return scanner.Err()
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
		log.Printf("Missing build info: buildID: \"%s\", jobName: \"%s\"\n", buildID, jobName)
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

	if err := c.checkRequirement(); err != nil {
		log.Printf("Failed to check requirement, %v", err)
		return 1 //Error code 1: Parse file failure
	}

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
		codeCoverage: make(map[string]float64),
		report:       *reportFile,
		requirement:  *requirementFile,
		bucket:       *gcsBucket,
	}

	os.Exit(c.checkPackageCoverage())
}

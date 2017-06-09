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

	"os"
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
	requirementFile = flag.String("requirement_file", "codecov.requirement", "Package code coverage requuirement.")
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
		fmt.Printf("Failed to open report file %s, %v", c.report, err)
		return err
	}
	defer func() {
		if err = f.Close(); err != nil {
			fmt.Printf("Failed to close file %s, %v", c.report, err)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) == 4 || parts[0] == "ok" {
			words := strings.Split(parts[3], " ")
			if len(words) == 4 {
				n, err := strconv.ParseFloat(strings.TrimSuffix(words[1], "%"), 64)
				if err != nil {
					fmt.Printf("Failed to parse coverage for package %s: %s, %v", words[1], parts[1], err)
					return err
				}
				c.codeCoverage[parts[1]] = n
			}
		}
	}

	return scanner.Err()
}

func (c *codecovChecker) checkRequirement() error {
	f, err := os.Open(c.requirement)
	if err != nil {
		fmt.Printf("Failed to open requirement file, %s, %v", c.requirement, err)
		return err
	}
	defer func() {
		if err = f.Close(); err != nil {
			fmt.Printf("Failed to close file %s, %s", c.requirement, err)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) == 2 {
			if cov, exist := c.codeCoverage[parts[0]]; exist {
				re, err := strconv.ParseFloat(parts[1], 64)
				if err != nil {
					fmt.Printf("Failed to get requirement for package %s: %s, %v", parts[0], parts[1], err)
					return err
				}
				if cov < re {
					c.failedPackage = append(c.failedPackage, fmt.Sprintf("%s\t%.2f\t%s", parts[0], cov, parts[1]))
				}

			} else {
				c.failedPackage = append(c.failedPackage, fmt.Sprintf("%s\t%s\t%s", parts[0], "0.0", parts[1]))
			}
		}
	}

	return scanner.Err()
}

func (c *codecovChecker) uploadCoverage() error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Printf("Failed to get storage client, %v", err)
		return err
	}

	jobName := os.Getenv(jobName)
	buildID := os.Getenv(buildID)
	if buildID == "" || jobName == "" {
		fmt.Printf("Missing build info: buildId: \"%s\", jobName: \"%s\"\n", buildID, jobName)
		return errors.New("missing build info")
	}

	object := jobName + "/" + buildID

	coverageString := ""
	for p, c := range c.codeCoverage {
		coverageString += fmt.Sprintf("%s\t%.2f\n", p, c)
	}

	w := client.Bucket(c.bucket).Object(object).NewWriter(ctx)
	if _, err = w.Write([]byte(coverageString)); err != nil {
		fmt.Printf("Failed to write coverage to gcs, %v", err)
	}

	defer func() {
		if err = w.Close(); err != nil {
			fmt.Printf("Failed to close gcs writer file, %v", err)
		}
	}()

	fmt.Printf("Successfully upload codecov.report %s", object)
	return nil
}

func main() {
	flag.Parse()

	c := &codecovChecker{
		codeCoverage: make(map[string]float64),
		report:       *reportFile,
		requirement:  *requirementFile,
		bucket:       *gcsBucket,
	}

	if err := c.parseReport(); err != nil {
		fmt.Printf("Failed to parse report")
		os.Exit(1)
	}

	if err := c.checkRequirement(); err != nil {
		fmt.Print("Failed to check requirement.")
		os.Exit(1)
	}

	if len(c.failedPackage) == 0 {
		fmt.Println("All packages passed code coverage requirements!")
	} else {
		fmt.Printf("Following package(s) failed to meet requirements:\nPackage Name\t\tActual Coverage\t\tRequirement\n")
		for _, p := range c.failedPackage {
			fmt.Println(p)
		}
		os.Exit(2)
	}

	if err := c.uploadCoverage(); err != nil {
		fmt.Print("Failed to upload coverage.")
		os.Exit(3)
	}
}

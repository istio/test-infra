// Copyright 2018 Istio Authors
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
	"flag"
	"log"

	ci2g "istio.io/test-infra/toolbox/ci_to_gubernator"
	u "istio.io/test-infra/toolbox/util"
)

const (
	unspecifiedInt = -1
	circleciBucket = "istio-circleci"
)

var (
	jobStarts   = flag.Bool("job_starts", false, "Mark the start of a job by creating started.json")
	jobFinishes = flag.Bool("job_finishes", false, "Mark the end of a job by creating finished.json")
	exitCode    = flag.Int("exit_code", unspecifiedInt, "Exit code returned from the test command")
	buildNum    = flag.Int("build_number", unspecifiedInt, "Build number genereated by CI")
	prNum       = flag.Int("pr_number", unspecifiedInt, "Pull request number on GitHub")
	sha         = flag.String("sha", "", "The commit from which the build and test were made")
	org         = flag.String("org", "", "Org of the GitHub project being built")
	repo        = flag.String("repo", "", "Repo of the GitHub project being built")
	job         = flag.String("job", "", "Name of job being built")
	junitXML    = flag.String("junit_xml", "", "Path to the junit xml report")
	buildLogTXT = flag.String("build_log_txt", "", "Path to the build log")
)

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if *jobStarts {
		createPushStartedJSON()
	} else if *jobFinishes {
		uploadArtifactsUpdateLatestBuild()
	} else {
		log.Fatalf("Either --job_starts or --job_finishes is required")
	}
}

func createPushStartedJSON() {
	u.AssertNotEmpty("sha", sha)
	u.AssertNotEmpty("org", org)
	u.AssertNotEmpty("repo", repo)
	u.AssertNotEmpty("job", job)
	u.AssertIntDefined("build_number", buildNum, unspecifiedInt)
	u.AssertIntDefined("pr_number", prNum, unspecifiedInt)
	cvt := ci2g.NewConverter(circleciBucket, *org, *repo, *job, *buildNum)
	if err := cvt.CreateUploadStartedJSON(*prNum, *sha); err != nil {
		log.Fatalf("Failed to create started.json: %v", err)
	}
}

func uploadArtifactsUpdateLatestBuild() {
}

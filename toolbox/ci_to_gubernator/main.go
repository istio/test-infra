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
)

var (
	jobStarts   = flag.Bool("job_starts", false, "Mark the start of a job by creating started.json")
	exitCode    = flag.Int("exit_code", unspecifiedInt, "Exit code returned from the test command")
	buildNum    = flag.Int("build_number", unspecifiedInt, "Build number genereated by CI")
	prNum       = flag.Int("pr_number", unspecifiedInt, "Pull request number on GitHub")
	sha         = flag.String("sha", "", "The commit from which the build and test were made")
	org         = flag.String("org", "", "Org of the GitHub project being built")
	repo        = flag.String("repo", "", "Repo of the GitHub project being built")
	junitXML    = flag.String("junit_xml", "", "Path to the junit xml report")
	buildLogTXT = flag.String("build_log_txt", "", "Path to the build log")
)

// func init() {
// 	flag.Parse()
// 	u.AssertIntDefined("exit_code", exitCode, unspecifiedInt)
// 	u.AssertIntDefined("build_num", buildNum, unspecifiedInt)
// 	u.AssertNotEmpty("sha", sha)
// 	u.AssertNotEmpty("org", org)
// 	u.AssertNotEmpty("repo", repo)
// 	u.AssertNotEmpty("junit_xml", junitXML)
// 	u.AssertNotEmpty("build_log_txt", buildLogTXT)
// }

func main() {
	if *jobStarts {
		createPushStartedJSON()
	}
	log.Printf("hello\n")
}

func createPushStartedJSON() {
	u.AssertNotEmpty("sha", sha)
	u.AssertNotEmpty("org", org)
	u.AssertNotEmpty("repo", repo)
	u.AssertIntDefined("pr_number", prNum, unspecifiedInt)
	if err := ci2g.CreateStartedJSON(*prNum); err != nil {
		log.Printf("Failed to create started.json")
	}
}

// Copyright 2019 Istio Authors
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
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"istio.io/test-infra/prow/config"
)

const ConfigOutput = "../../cluster/jobs"

func exit(err error, context string) {
	if context == "" {
		_, _ = fmt.Fprint(os.Stderr, fmt.Sprintf("%v", err))
	} else {
		_, _ = fmt.Fprint(os.Stderr, fmt.Sprintf("%v: %v", context, err))
	}
	os.Exit(1)
}

func GetFileName(repo string, org string, branch string) string {
	key := fmt.Sprintf("%s.%s.%s.gen.yaml", org, repo, branch)
	return path.Join(ConfigOutput, org, repo, key)
}

func main() {
	if len(os.Args) != 2 {
		panic("must provide one of write, diff, check, print")
	}
	files, err := ioutil.ReadDir("../jobs")
	if err != nil {
		exit(err, "failed to read jobs")
	}
	for _, file := range files {
		jobs := config.ReadJobConfig(path.Join("..", "jobs", file.Name()))
		for _, branch := range jobs.Branches {
			config.ValidateJobConfig(jobs)
			output := config.ConvertJobConfig(jobs, branch)
			fname := GetFileName(jobs.Repo, jobs.Org, branch)
			switch os.Args[1] {
			case "write":
				config.WriteConfig(output, fname)
			case "diff":
				existing := config.ReadProwJobConfig(fname)
				config.DiffConfig(output, existing)
			case "check":
				if err := config.CheckConfig(output, fname); err != nil {
					exit(err, "check failed")
				}
			default:
				config.PrintConfig(output)
			}
		}
	}
}

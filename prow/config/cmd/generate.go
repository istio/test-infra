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
	"path/filepath"

	"istio.io/test-infra/prow/config"
)

const ConfigOutput = "../../cluster/jobs"

func exit(err error, context string) {
	if context == "" {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "%v: %v\n", context, err)
	}
	os.Exit(1)
}

func GetFileName(repo string, org string, branch string) string {
	key := fmt.Sprintf("%s.%s.%s.gen.yaml", org, repo, branch)
	return path.Join(ConfigOutput, org, repo, key)
}

func main() {

	// TODO: deserves a better CLI...
	if len(os.Args) < 2 {
		panic("must provide one of write, diff, print, branch")
	} else if os.Args[1] == "branch" {
		if len(os.Args) != 3 {
			panic("must specify branch name")
		}
	} else if len(os.Args) != 2 {
		panic("too many arguments")
	}

	files, err := ioutil.ReadDir("../jobs")
	if err != nil {
		exit(err, "failed to read jobs")
	}

	if os.Args[1] == "branch" {
		for _, file := range files {
			src := path.Join("..", "jobs", file.Name())

			jobs := config.ReadJobConfig(src)
			if jobs.SupportReleaseBranching {
				jobs.Branches = []string{"release-" + os.Args[2]}
				jobs.SupportReleaseBranching = false

				name := file.Name()
				ext := filepath.Ext(name)
				name = name[:len(name)-len(ext)] + "-" + os.Args[2] + ext

				dst := path.Join("..", "jobs", name)
				if err := config.WriteJobConfig(jobs, dst); err != nil {
					exit(err, "writing branched config failed")
				}
			}
		}
	} else {
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
				default:
					config.PrintConfig(output)
				}
			}
		}
	}
}

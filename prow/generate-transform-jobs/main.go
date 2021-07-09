// Copyright Istio Authors
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
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"istio.io/test-infra/prow/genjobs/pkg/configuration"
)

func exit(err error, context string) {
	if context == "" {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "%v: %v\n", context, err)
	}
	os.Exit(1)
}

var inputDir = flag.String("input-dir", "../config/istio-private_jobs", "directory of input jobs")

// branchJobSlices updates transform jobs slices such as allow and deny jobs using a branch name
func branchJobSlices(in []string, branch string) []string {
	for key, val := range in {
		if strings.HasSuffix(val, "_postsubmit") {
			val = strings.Replace(val, "_postsubmit", fmt.Sprintf("_%s_postsubmit", branch), 1)
		} else if strings.HasSuffix(val, "_presubmit") {
			val = strings.Replace(val, "_presubmit", fmt.Sprintf("_%s_presubmit", branch), 1)
		} else {
			val = fmt.Sprintf("%s_%s", val, branch)
		}
		in[key] = val
	}
	return in
}

// Note that this app mirrors the functionality of prow/cmd/generate.go, but acting on transformations instead of prow jobs.
// Any changes made here should also be considered for prow/cmd/generate.go.
func main() {
	flag.Parse()

	// TODO: deserves a better CLI...
	if len(flag.Args()) < 1 || flag.Arg(0) != "branch" {
		panic("Branch is the only supported operation. Diff and print are not implemented. Write is handled by genjobs")
	} else if flag.Arg(0) == "branch" {
		if len(flag.Args()) != 2 {
			panic("must specify branch name")
		}
	} else if len(flag.Args()) != 1 {
		panic("too many arguments")
	}

	if os.Args[1] == "branch" {
		if err := filepath.Walk(*inputDir, func(src string, file os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("error: %s\n", err.Error())
			}

			if file.IsDir() {
				return nil
			}
			if filepath.Ext(file.Name()) != ".yaml" && filepath.Ext(file.Name()) != ".yml" || file.Name() == ".global.yaml" {
				log.Println("skipping", file.Name())
				return nil
			}
			jobs := configuration.ReadTransformJobsConfig(src)
			if jobs.SupportReleaseBranching {
				branch := "release-" + flag.Arg(1)
				jobs.Defaults.Branches = []string{branch}
				jobs.SupportReleaseBranching = false
				jobs.Defaults.Modifier = strings.Replace(jobs.Defaults.Modifier, "master_", fmt.Sprintf("%s_", branch), 1)

				for key, transform := range jobs.Transforms {
					transform.JobAllowlist = branchJobSlices(transform.JobAllowlist, branch)
					transform.JobDenylist = branchJobSlices(transform.JobDenylist, branch)

					for key, val := range transform.Labels {
						transform.Labels[key] = strings.Replace(val, "master", branch, 1)
					}

					jobs.Transforms[key] = transform

				}
				name := file.Name()
				ext := filepath.Ext(name)
				name = name[:len(name)-len(ext)] + "-" + flag.Arg(1) + ext

				dst := path.Join(*inputDir, name)
				if err := configuration.WriteTransformJobConfig(jobs, dst); err != nil {
					exit(err, "writing branched config failed")
				}
			}

			return nil
		}); err != nil {
			exit(err, "walking through the private meta config files failed")
		}
	} else {
		// may be useful to add the print and diff functionality here.
		exit(nil, "other operations are currently not supported by this utility. Please see cmd/generate.go or genjobs")
	}
}

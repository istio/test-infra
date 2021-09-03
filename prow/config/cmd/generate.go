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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"

	k8sProwConfig "k8s.io/test-infra/prow/config"

	"istio.io/test-infra/prow/config"
)

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
	return path.Join(*outputDir, org, repo, key)
}

var (
	inputDir  = flag.String("input-dir", "../jobs", "directory of input jobs")
	outputDir = flag.String("output-dir", "../../cluster/jobs", "directory of output jobs")
)

func main() {
	flag.Parse()

	// TODO: deserves a better CLI...
	if len(flag.Args()) < 1 {
		panic("must provide one of write, diff, print, branch")
	} else if flag.Arg(0) == "branch" {
		if len(flag.Args()) != 2 {
			panic("must specify branch name")
		}
	} else if len(flag.Args()) != 1 {
		panic("too many arguments")
	}

	var bc *config.BaseConfig
	if _, err := os.Stat(filepath.Join(*inputDir, ".base.yaml")); !os.IsNotExist(err) {
		bc = config.ReadBase(nil, filepath.Join(*inputDir, ".base.yaml"))
	}

	if os.Args[1] == "branch" {
		if err := filepath.WalkDir(*inputDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("error: %s\n", err.Error())
			}
			baseConfig := bc
			if _, err := os.Stat(filepath.Join(path, ".base.yaml")); !os.IsNotExist(err) {
				baseConfig = config.ReadBase(baseConfig, filepath.Join(path, ".base.yaml"))
			}
			cli := config.Client{BaseConfig: baseConfig}

			files, _ := ioutil.ReadDir(path)
			for _, file := range files {
				if file.IsDir() {
					continue
				}

				if (filepath.Ext(file.Name()) != ".yaml" && filepath.Ext(file.Name()) != ".yml") ||
					file.Name() == ".base.yaml" {
					log.Println("skipping", file.Name())
					continue
				}

				src := filepath.Join(path, file.Name())
				jobs := cli.ReadJobsConfig(src)
				jobs.Jobs = config.FilterReleaseBranchingJobs(jobs.Jobs)

				if jobs.SupportReleaseBranching {
					tagRegex := regexp.MustCompile(`^(.+):(.+)-([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}-[0-9]{2}-[0-9]{2})$`)
					match := tagRegex.FindStringSubmatch(jobs.Image)
					branch := "release-" + flag.Arg(1)
					if len(match) == 4 {
						newImage := fmt.Sprintf("%s:%s-%s", match[1], branch, match[3])
						if err := exec.Command("gcloud", "container", "images", "add-tag", match[0], newImage).Run(); err != nil {
							exit(err, "unable to add image tag: "+newImage)
						} else {
							jobs.Image = newImage
						}
					}
					jobs.Branches = []string{branch}
					jobs.SupportReleaseBranching = false

					name := file.Name()
					ext := filepath.Ext(name)
					name = name[:len(name)-len(ext)] + "-" + flag.Arg(1) + ext

					dst := filepath.Join(*inputDir, name)
					if err := config.WriteJobConfig(jobs, dst); err != nil {
						exit(err, "writing branched config failed")
					}
				}
			}

			return nil
		}); err != nil {
			exit(err, "walking through the meta config files failed")
		}
	} else {
		type ref struct {
			org    string
			repo   string
			branch string
		}
		// Store the job config generated from all meta-config files in a cache map, and combine the
		// job configs before we generate the final config files.
		// In this way we can have multiple meta-config files for the same org/repo:branch
		cachedOutput := map[ref]k8sProwConfig.JobConfig{}
		if err := filepath.WalkDir(*inputDir, func(path string, d os.DirEntry, err error) error {
			if !d.IsDir() {
				return nil
			}
			if err != nil {
				fmt.Printf("error: %s\n", err.Error())
			}

			baseConfig := bc
			if _, err := os.Stat(filepath.Join(path, ".base.yaml")); !os.IsNotExist(err) {
				baseConfig = config.ReadBase(baseConfig, filepath.Join(path, ".base.yaml"))
			}
			cli := config.Client{BaseConfig: baseConfig}

			files, _ := ioutil.ReadDir(path)
			for _, file := range files {
				if file.IsDir() {
					continue
				}

				if (filepath.Ext(file.Name()) != ".yaml" && filepath.Ext(file.Name()) != ".yml") ||
					file.Name() == ".base.yaml" {
					log.Println("skipping", file.Name())
					continue
				}

				src := filepath.Join(path, file.Name())
				jobs := cli.ReadJobsConfig(src)
				for _, branch := range jobs.Branches {
					cli.ValidateJobConfig(file.Name(), jobs)
					output := cli.ConvertJobConfig(jobs, branch)
					rf := ref{jobs.Org, jobs.Repo, branch}
					if _, ok := cachedOutput[rf]; !ok {
						cachedOutput[rf] = output
					} else {
						cachedOutput[rf] = combineJobConfigs(cachedOutput[rf], output,
							fmt.Sprintf("%s/%s", jobs.Org, jobs.Repo))
					}
				}
			}
			return nil
		}); err != nil {
			exit(err, "walking through the meta config files failed")
		}

		for r, output := range cachedOutput {
			fname := GetFileName(r.repo, r.org, r.branch)
			switch flag.Arg(0) {
			case "write":
				config.Write(output, fname, bc.AutogenHeader)
			case "diff":
				existing := config.ReadProwJobConfig(fname)
				config.Diff(output, existing)
			default:
				config.Print(output)
			}
		}
	}
}

func combineJobConfigs(jc1, jc2 k8sProwConfig.JobConfig, orgRepo string) k8sProwConfig.JobConfig {
	presubmits := jc1.PresubmitsStatic
	postsubmits := jc1.PostsubmitsStatic
	periodics := jc1.Periodics

	presubmits[orgRepo] = append(presubmits[orgRepo], jc2.PresubmitsStatic[orgRepo]...)
	postsubmits[orgRepo] = append(postsubmits[orgRepo], jc2.PostsubmitsStatic[orgRepo]...)
	periodics = append(periodics, jc2.Periodics...)

	return k8sProwConfig.JobConfig{
		PresubmitsStatic:  presubmits,
		PostsubmitsStatic: postsubmits,
		Periodics:         periodics,
	}
}

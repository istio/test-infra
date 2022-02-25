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

	"github.com/hashicorp/go-multierror"
	shell "github.com/kballard/go-shellquote"
	k8sProwConfig "k8s.io/test-infra/prow/config"
	"sigs.k8s.io/yaml"

	"istio.io/test-infra/tools/prowgen/pkg"
	"istio.io/test-infra/tools/prowgen/pkg/spec"
)

var (
	// regex to match the test image tags.
	tagRegex = regexp.MustCompile(`^(.+):(.+)-([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}-[0-9]{2}-[0-9]{2})$`)

	inputDir            = flag.String("input-dir", "./prow/config/jobs", "directory of input jobs")
	outputDir           = flag.String("output-dir", "./prow/cluster/jobs", "directory of output jobs")
	preprocessCommand   = flag.String("pre-process-command", "", "command to run to preprocess the meta config files")
	postprocessCommand  = flag.String("post-process-command", "", "command to run to postprocess the generated config files")
	longJobNamesAllowed = flag.Bool("allow-long-job-names", false, "allow job names that are longer than 63 characters")
)

func main() {
	flag.Parse()

	// TODO: deserves a better CLI...
	if len(flag.Args()) < 1 {
		panic("must provide one of write, print, check, branch")
	} else if flag.Arg(0) == "branch" {
		if len(flag.Args()) != 2 {
			panic("must specify branch name")
		}
	} else if len(flag.Args()) != 1 {
		panic("too many arguments")
	}

	var bc spec.BaseConfig
	if _, err := os.Stat(filepath.Join(*inputDir, ".base.yaml")); !os.IsNotExist(err) {
		bc = pkg.ReadBase(nil, filepath.Join(*inputDir, ".base.yaml"))
	}

	if os.Args[1] == "branch" {
		if err := filepath.WalkDir(*inputDir, func(path string, d os.DirEntry, err error) error {
			if !d.IsDir() {
				return nil
			}
			if err != nil {
				log.Fatal(err)
			}
			baseConfig := bc
			if _, err := os.Stat(filepath.Join(path, ".base.yaml")); !os.IsNotExist(err) {
				baseConfig = pkg.ReadBase(&baseConfig, filepath.Join(path, ".base.yaml"))
			}
			cli := pkg.Client{BaseConfig: baseConfig, LongJobNamesAllowed: *longJobNamesAllowed}

			files, _ := ioutil.ReadDir(path)
			for _, file := range files {
				if file.IsDir() {
					continue
				}

				if (filepath.Ext(file.Name()) != ".yaml" && filepath.Ext(file.Name()) != ".yml") ||
					file.Name() == ".base.yaml" {
					log.Println("skipping non-yaml file: ", file.Name())
					continue
				}

				src := filepath.Join(path, file.Name())
				jobs := cli.ReadJobsConfig(src)
				jobs.Jobs = pkg.FilterReleaseBranchingJobs(jobs.Jobs)

				if jobs.SupportReleaseBranching {
					match := tagRegex.FindStringSubmatch(jobs.Image)
					branch := "release-" + flag.Arg(1)
					if len(match) == 4 {
						// HACK: replacing the branch name in the image tag and
						// adding it as a new tag.
						// For example, if the test image in the current Prow job
						// config is
						// `gcr.io/istio-testing/build-tools:release-1.10-2021-08-09T16-46-08`,
						// and the Prow job config for release-1.11 branch is
						// supposed to be generated, the image will be added a
						// new `release-1.11-2021-08-09T16-46-08` tag.
						// This is only needed for creating Prow jobs for a new
						// release branch for the first time, and the image tag
						// will be overwritten by Automator the next time the
						// image for the new branch is updated.
						newImage := fmt.Sprintf("%s:%s-%s", match[1], branch, match[3])
						if err := exec.Command("gcloud", "container", "images", "add-tag", match[0], newImage).Run(); err != nil {
							log.Fatalf("Unable to add image tag %q: %v", newImage, err)
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
					bytes, err := yaml.Marshal(jobs)
					if err != nil {
						log.Fatalf("Error marshaling jobs config: %v", err)
					}

					// Writes the job yaml
					if err := ioutil.WriteFile(dst, bytes, 0o644); err != nil {
						log.Fatalf("Error writing branches config: %v", err)
					}
				}
			}

			return nil
		}); err != nil {
			log.Fatalf("Walking through the meta config files failed: %v", err)
		}
	} else {
		if *preprocessCommand != "" {
			if err := runProcessCommand(*preprocessCommand); err != nil {
				log.Fatalf("Error running preprocess command %q: %v", *preprocessCommand, err)
			}
		}

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
				log.Fatal(err)
			}

			baseConfig := bc
			if _, err := os.Stat(filepath.Join(path, ".base.yaml")); !os.IsNotExist(err) {
				baseConfig = pkg.ReadBase(&baseConfig, filepath.Join(path, ".base.yaml"))
			}
			cli := pkg.Client{BaseConfig: baseConfig, LongJobNamesAllowed: *longJobNamesAllowed}

			files, _ := ioutil.ReadDir(path)
			for _, file := range files {
				if file.IsDir() {
					continue
				}

				if (filepath.Ext(file.Name()) != ".yaml" && filepath.Ext(file.Name()) != ".yml") ||
					file.Name() == ".base.yaml" {
					log.Println("skipping non-yaml file: ", file.Name())
					continue
				}

				src := filepath.Join(path, file.Name())
				jobs := cli.ReadJobsConfig(src)
				for _, branch := range jobs.Branches {
					output, err := cli.ConvertJobConfig(file.Name(), jobs, branch)
					if err != nil {
						log.Fatal(err)
					}
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
			log.Fatalf("Walking through the meta config files failed: %v", err)
		}

		var err error
		for r, output := range cachedOutput {
			fname := outputFileName(r.repo, r.org, r.branch)
			switch flag.Arg(0) {
			case "write":
				if e := pkg.Write(output, fname, bc.AutogenHeader); e != nil {
					err = multierror.Append(err, e)
				}
				if *postprocessCommand != "" {
					if e := runProcessCommand(*postprocessCommand); e != nil {
						err = multierror.Append(err, e)
					}
				}
			case "check":
				if e := pkg.Check(output, fname, bc.AutogenHeader); e != nil {
					err = multierror.Append(err, e)
				}
			case "print":
				pkg.Print(output)
			}
		}

		if err != nil {
			log.Fatalf("Get errors for the %q operation:\n%v", flag.Arg(0), err)
		}
	}
}

func runProcessCommand(rawCommand string) error {
	log.Printf("⚙️ %s", rawCommand)
	cmdSplit, err := shell.Split(rawCommand)
	if len(cmdSplit) == 0 || err != nil {
		return fmt.Errorf("error parsing the command %q: %w", rawCommand, err)
	}
	cmd := exec.Command(cmdSplit[0], cmdSplit[1:]...)

	// Set INPUT and OUTPUT env vars for the pre-process and post-process
	// commands to consume.
	cmd.Env = append(os.Environ(), "INPUT="+*inputDir, "OUTPUT="+*outputDir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func outputFileName(repo string, org string, branch string) string {
	key := fmt.Sprintf("%s.%s.%s.gen.yaml", org, repo, branch)
	return path.Join(*outputDir, org, repo, key)
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

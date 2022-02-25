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

package pkg

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/google/go-cmp/cmp"
	"k8s.io/test-infra/prow/config"
	"sigs.k8s.io/yaml"

	"istio.io/test-infra/tools/prowgen/pkg/spec"
)

// WriteJobsConfig will write the meta jobs to the given file.
func WriteJobsConfig(jobsConfig spec.JobsConfig, file string) error {
	bytes, err := yaml.Marshal(jobsConfig)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(file, bytes, 0o644)
}

// Write will write the generated Prow jobs to the given file.
func Write(jobs config.JobConfig, fname, header string) error {
	bs, err := yaml.Marshal(jobs)
	if err != nil {
		log.Fatalf("Failed to marshal result: %v", err)
	}
	dir := filepath.Dir(fname)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalf("failed to create directory %q: %v", dir, err)
	}
	if header == "" {
		header = DefaultAutogenHeader
	}
	output := []byte(header + "\n")
	output = append(output, bs...)
	return ioutil.WriteFile(fname, output, 0o644)
}

// Check will diff the generated config file and the current config file.
func Check(jobs config.JobConfig, currentConfigFile string, header string) error {
	current, err := ioutil.ReadFile(currentConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read current config for %s: %v", currentConfigFile, err)
	}

	newConfig, err := yaml.Marshal(jobs)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}
	if header == "" {
		header = DefaultAutogenHeader
	}
	output := []byte(header + "\n")
	output = append(output, newConfig...)

	if diff := cmp.Diff(output, current); diff != "" {
		return fmt.Errorf("generated config is different from file %s\nWant(-), got(+):\n%s", currentConfigFile, diff)
	}
	return nil
}

// Print will print out the generated Prow jobs config.
func Print(jobs config.JobConfig) {
	bs, err := yaml.Marshal(jobs)
	if err != nil {
		log.Fatalf("Failed to write result: %v", err)
	}
	fmt.Println(string(bs))
}

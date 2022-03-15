/*
Copyright Istio Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package configuration

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	prowjob "k8s.io/test-infra/prow/apis/prowjobs/v1"

	"istio.io/test-infra/tools/prowtrans/pkg/util"
)

// Configuration is the yaml configuration file format.
type Configuration struct {
	Org                     string      `json:"org,omitempty"`
	Repo                    string      `json:"repo,omitempty"`
	SupportReleaseBranching bool        `json:"support_release_branching,omitempty"`
	Defaults                Transform   `json:"defaults,omitempty"`
	Transforms              []Transform `json:"transforms,omitempty"`
}

// transform are the available transformation fields.
type Transform struct {
	Annotations            map[string]string       `json:"annotations,omitempty"`
	Bucket                 string                  `json:"bucket,omitempty"`
	Cluster                string                  `json:"cluster,omitempty"`
	Channel                string                  `json:"channel,omitempty"`
	SSHKeySecret           string                  `json:"ssh-key-secret,omitempty"`
	Modifier               string                  `json:"modifier,omitempty"`
	ServiceAccount         string                  `json:"service_account_name,omitempty"`
	Input                  string                  `json:"input,omitempty"`
	Output                 string                  `json:"output,omitempty"`
	Sort                   string                  `json:"sort,omitempty"`
	ExtraRefs              []prowjob.Refs          `json:"extra-refs,omitempty"`
	ReporterConfig         *prowjob.ReporterConfig `json:"reporter_config,omitempty"`
	Branches               []string                `json:"branches,omitempty"`
	BranchesOut            []string                `json:"branches-out,omitempty"`
	RefBranchOut           string                  `json:"ref-branch-out,omitempty"`
	Presets                []string                `json:"presets,omitempty"`
	RerunOrgs              []string                `json:"rerun-orgs,omitempty"`
	RerunUsers             []string                `json:"rerun-users,omitempty"`
	EnvDenylist            []string                `json:"env-denylist,omitempty"`
	VolumeDenylist         []string                `json:"volume-denylist,omitempty"`
	JobAllowlist           []string                `json:"job-allowlist,omitempty"`
	JobDenylist            []string                `json:"job-denylist,omitempty"`
	RepoAllowlist          []string                `json:"repo-allowlist,omitempty"`
	RepoDenylist           []string                `json:"repo-denylist,omitempty"`
	JobType                []string                `json:"job-type,omitempty"`
	Selector               map[string]string       `json:"selector,omitempty"`
	Labels                 map[string]string       `json:"labels,omitempty"`
	Env                    map[string]string       `json:"env,omitempty"`
	RefOrgMap              map[string]string       `json:"ref-mapping,omitempty"`
	OrgMap                 map[string]string       `json:"mapping,omitempty"`
	HubMap                 map[string]string       `json:"hub,omitempty"`
	Tag                    string                  `json:"tag,omitempty"`
	Clean                  bool                    `json:"clean,omitempty"`
	DryRun                 bool                    `json:"dry-run,omitempty"`
	Refs                   bool                    `json:"refs,omitempty"`
	Resolve                bool                    `json:"resolve,omitempty"`
	SSHClone               bool                    `json:"ssh-clone,omitempty"`
	OverrideSelector       bool                    `json:"override-selector,omitempty"`
	SupportGerritReporting bool                    `json:"support-gerrit-reporting,omitempty"`
	AllowLongJobNames      bool                    `json:"allow-long-job-names,omitempty"`
	Verbose                bool                    `json:"verbose,omitempty"`
}

// ReadTransformJobsConfig reads the private jobs yaml
func ReadTransformJobsConfig(file string) Configuration {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		util.PrintErrAndExit(fmt.Errorf("failed to read %s", file))
	}

	jobsConfig := Configuration{}
	if err := yaml.Unmarshal(yamlFile, &jobsConfig); err != nil {
		util.PrintErrAndExit(fmt.Errorf("failed to unmarshal %s", file))
	}

	return jobsConfig
}

// WriteTransformJobConfig writes the job yaml
func WriteTransformJobConfig(jobsConfig Configuration, file string) error {
	bytes, err := yaml.Marshal(jobsConfig)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(file, bytes, 0o644)
}

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

package spec

import (
	"log"

	v1 "k8s.io/api/core/v1"
	prowjob "k8s.io/test-infra/prow/apis/prowjobs/v1"
	"sigs.k8s.io/yaml"
)

// BaseConfig represents the fields that can be defined in a .base.yaml file,
// which is shared by all the meta job config files under the same folder.
type BaseConfig struct {
	CommonConfig

	AutogenHeader string `json:"autogen_header,omitempty"`

	PathAliases map[string]string `json:"path_aliases,omitempty"`

	ClusterOverrides map[string]string `json:"cluster_overrides,omitempty"`

	TestgridConfig TestgridConfig `json:"testgrid_config,omitempty"`
}

func (baseConfig *BaseConfig) DeepCopy() BaseConfig {
	bc, _ := yaml.Marshal(baseConfig)
	newBaseConfig := BaseConfig{}
	if err := yaml.Unmarshal(bc, &newBaseConfig); err != nil {
		log.Fatalf("Failed to unmarshal BaseConfig: %v", err)
	}
	return newBaseConfig
}

type TestgridConfig struct {
	Enabled            bool   `json:"enabled,omitempty"`
	AlertEmail         string `json:"alert_email,omitempty"`
	NumFailuresToAlert string `json:"num_failures_to_alert,omitempty"`
}

// JobsConfig represents the fields that can be defined in a meta job file, and
// it can contain multiple Jobs.
type JobsConfig struct {
	CommonConfig

	SupportReleaseBranching bool `json:"support_release_branching,omitempty"`

	Repo     string   `json:"repo,omitempty"`
	Org      string   `json:"org,omitempty"`
	CloneURI string   `json:"clone_uri,omitempty"`
	Branches []string `json:"branches,omitempty"`

	Jobs []Job `json:"jobs,omitempty"`
}

// Job is the last layer for defining the actual Prow jobs.
type Job struct {
	CommonConfig

	DisableReleaseBranching bool `json:"disable_release_branching,omitempty"`

	Name    string   `json:"name,omitempty"`
	Command []string `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	Types   []string `json:"types,omitempty"`
	Repos   []string `json:"repos,omitempty"`
	// Architectures defines architectures to build as. Defaults to amd64.
	Architectures []string `json:"architectures,omitempty"`

	GerritPresubmitLabel  string `json:"gerrit_presubmit_label,omitempty"`
	GerritPostsubmitLabel string `json:"gerrit_postsubmit_label,omitempty"`

	ReporterConfig *prowjob.ReporterConfig `json:"reporter_config,omitempty"`
}

// CommonConfig contains all the common fields that can be overlayed through
// BaseConfig->JobsConfig->Job
type CommonConfig struct {
	GCSLogBucket                  string `json:"gcs_log_bucket,omitempty"`
	TerminationGracePeriodSeconds int64  `json:"termination_grace_period_seconds,omitempty"`

	Interval string `json:"interval,omitempty"`
	Cron     string `json:"cron,omitempty"`

	Cluster      string            `json:"cluster,omitempty"`
	NodeSelector map[string]string `json:"node_selector,omitempty"`

	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`

	Matrix map[string][]string `json:"matrix,omitempty"`
	Params map[string]string   `json:"params,omitempty"`

	ResourcePresets      map[string]v1.ResourceRequirements `json:"resources_presets,omitempty"`
	RequirementPresets   map[string]RequirementPreset       `json:"requirement_presets,omitempty"`
	Requirements         []string                           `json:"requirements,omitempty"`
	ExcludedRequirements []string                           `json:"excluded_requirements,omitempty"`

	Env                []v1.EnvVar `json:"env,omitempty"`
	Image              string      `json:"image,omitempty"`
	ImagePullPolicy    string      `json:"image_pull_policy,omitempty"`
	ImagePullSecrets   []string    `json:"image_pull_secrets,omitempty"`
	ServiceAccountName string      `json:"service_account_name,omitempty"`

	Regex   string `json:"regex,omitempty"`
	Trigger string `json:"trigger,omitempty"`

	Timeout        *prowjob.Duration `json:"timeout,omitempty"`
	MaxConcurrency int               `json:"max_concurrency,omitempty"`

	Resources string   `json:"resources,omitempty"`
	Modifiers []string `json:"modifiers,omitempty"`
}

func (commonConfig *CommonConfig) DeepCopy() CommonConfig {
	cc, _ := yaml.Marshal(commonConfig)
	newCommonConfig := CommonConfig{}
	if err := yaml.Unmarshal(cc, &newCommonConfig); err != nil {
		log.Fatalf("Failed to unmarshal CommonConfig: %v", err)
	}
	return newCommonConfig
}

// RequirementPreset can be used to re-use settings across multiple jobs.
type RequirementPreset struct {
	Annotations  map[string]string `json:"annotations,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Env          []v1.EnvVar       `json:"env,omitempty"`
	Volumes      []v1.Volume       `json:"volumes,omitempty"`
	VolumeMounts []v1.VolumeMount  `json:"volumeMounts,omitempty"`
	Args         []string          `json:"args,omitempty"`
}

func (r *RequirementPreset) DeepCopy() RequirementPreset {
	rp, _ := yaml.Marshal(r)
	newRequirementPreset := RequirementPreset{}
	if err := yaml.Unmarshal(rp, &newRequirementPreset); err != nil {
		log.Fatalf("Failed to unmarshal RequirementPreset: %v", err)
	}
	return newRequirementPreset
}

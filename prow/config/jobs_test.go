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

package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/test-infra/prow/config"
)

var (
	PrivateClusters = sets.NewString("private", "prow-arm-private")
	PublicClusters  = sets.NewString("default", "prow-arm", "test-infra-trusted")
)

func TestJobs(t *testing.T) {
	RunTest := BuildRunTest(t)

	RunTest("tests use correct cluster", func(j Job) error {
		switch j.Org() {
		case "istio-private":
			if !PrivateClusters.Has(j.Base.Cluster) {
				return fmt.Errorf("private org must use private cluster, got %v", j.Base.Cluster)
			}
		case "istio":
			if !PublicClusters.Has(j.Base.Cluster) {
				return fmt.Errorf("periodic run on unexpected cluster: %v", j.Base.Cluster)
			}
		default:
			if j.Type != Periodic {
				return fmt.Errorf("unknown org: %v", j.Org())
			}
			if !PublicClusters.Has(j.Base.Cluster) {
				return fmt.Errorf("periodic run on unexpected cluster: %v", j.Base.Cluster)
			}
		}
		return nil
	})

	RunTest("only secure jobs use trusted cluster", func(j Job) error {
		if j.Base.Cluster != "test-infra-trusted" {
			return nil
		}
		if j.Type == Presubmit {
			return fmt.Errorf("trusted jobs cannot run in presubmit")
		}
		return nil
	})

	RunTest("secure jobs do not use insecure caches", func(j Job) error {
		if j.Base.Cluster != "test-infra-trusted" {
			return nil
		}
		if j.Type == Presubmit {
			return fmt.Errorf("trusted jobs cannot run in presubmit")
		}
		return nil
	})

	// check to make sure we did not miss any volumes. This may just mean we need to update the test.
	RunTest("known volumes only", func(j Job) error {
		unknown := j.Volumes().Difference(AllVolumes)
		if len(unknown) == 0 {
			return nil
		}
		return fmt.Errorf("unknown volume type: %v", unknown.UnsortedList())
	})
	RunTest("presubmit jobs do not use privileged volumes", func(j Job) error {
		if j.Type != Presubmit {
			return nil
		}
		// Private volumes are handled in another test
		priv := j.Volumes().Difference(LowPrivilegeVolumes).Difference(PrivateVolumes)
		if len(priv) == 0 {
			return nil
		}
		return fmt.Errorf("presubmit job using privileged volume: %v", priv.UnsortedList())
	})
	RunTest("private volumes only used in private jobs", func(j Job) error {
		private := j.Org() == "istio-private"
		usesPrivate := j.Volumes().Intersection(PrivateVolumes).Len() > 0
		if usesPrivate && !private {
			return fmt.Errorf("only private jobs can use private volumes")
		}
		return nil
	})
	RunTest("release volumes only used in release jobs", func(j Job) error {
		releaseJob := j.RepoOrg == "istio/release-builder" && j.Type == Postsubmit
		// TODO: these shouldn't need grafana or docker, and they actually don't - the private cluster has empty secrets
		privateReleaseJob := j.RepoOrg == "istio-private/release-builder" && j.Type == Postsubmit
		baseImageBuilder := ((j.RepoOrg == "istio/istio" && j.Type == Postsubmit) || j.Type == Periodic) && j.BaseName() == "build-base-images"
		if releaseJob || privateReleaseJob || baseImageBuilder {
			return nil
		}
		usesReleaseVolumes := j.Volumes().Intersection(ReleaseVolumes).Len() > 0
		if usesReleaseVolumes {
			return fmt.Errorf("only release jobs can use release volumes, found %v", j.Volumes().Intersection(ReleaseVolumes).UnsortedList())
		}
		return nil
	})

	RunTest("service accounts", func(j Job) error {
		s, f := ServiceAccounts[j.ServiceAccount()]
		if !f {
			return fmt.Errorf("unknown service account: %q", j.ServiceAccount())
		}
		switch s {
		case LowPrivilege:
			// Anyone can use low privilege accounts
			return nil
		case MediumPrivilege:
			// Postsubmit job can use
			if j.Type != Presubmit {
				return nil
			}
			// Only proxy is allowed to run these jobs, which use RBE.
			if j.ServiceAccount() == "prowjob-rbe" && j.Repo() == "proxy" {
				return nil
			}
			if j.ServiceAccount() == "prowjob-github-read" && strings.HasPrefix(j.Name, "release-notes") {
				// Only release notes job is allowed
				return nil
			}
			return fmt.Errorf("privileged service account %v cannot run as presubmit", j.ServiceAccount())
		case HighPrivilege:
			legacyJob := strings.HasPrefix(j.Name, "dry-run_release-builder")
			if !legacyJob && j.Type == Presubmit {
				return fmt.Errorf("privileged service accounts cannot run as presubmit")
			}
		default:
			return fmt.Errorf("unknown sensitivity: %v", s)
		}

		return nil
	})
	RunTest("private service account only used in private jobs", func(j Job) error {
		private := j.Org() == "istio-private"
		usesPrivate := PrivateServiceAccounts.Has(j.ServiceAccount())
		if usesPrivate && !private {
			return fmt.Errorf("only private jobs can use private service account %q", j.ServiceAccount())
		}
		return nil
	})

	RunTest("selectors", func(j Job) error {
		// Node selectors are not used on trusted cluster (for now, anyways)
		if j.Base.Cluster == "test-infra-trusted" {
			return nil
		}
		validSelectors := []map[string]string{}
		for _, arch := range []string{"amd64", "arm64"} {
			for _, tpe := range []string{"test-pool", "build-pool"} {
				validSelectors = append(validSelectors, map[string]string{
					"kubernetes.io/arch": arch,
					"testing":            tpe,
				})
			}
		}
		ns := j.Base.Spec.NodeSelector
		for _, s := range validSelectors {
			if maps.Equal(s, ns) {
				// It's a known selector
				return nil
			}
		}
		return fmt.Errorf("unexpected node selector: %+v", ns)
	})

	RunTest("resources", func(j Job) error {
		// Resource requests are not used (for now) on trusted cluster
		if j.Base.Cluster == "test-infra-trusted" {
			return nil
		}
		for _, c := range j.Base.Spec.Containers {
			r := c.Resources
			if r.Requests.Cpu().IsZero() {
				return fmt.Errorf("cpu requests should be set")
			}
			if r.Requests.Memory().IsZero() {
				return fmt.Errorf("memory requests should be set")
			}
		}
		return nil
	})

	RunTest("container build", func(j Job) error {
		for _, c := range j.Base.Spec.Containers {
			if !strings.HasPrefix(c.Name, "gcr.io/istio-testing/build-tools") {
				continue
			}
			found := false
			for _, e := range c.Env {
				if e.Name == "BUILD_WITH_CONTAINER" && e.Value == "0" {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("must set BUILD_WITH_CONTAINER=0 to avoid nested containers")
			}
		}
		return nil
	})

	RunTest("token mount", func(j Job) error {
		st := j.Base.Spec.AutomountServiceAccountToken
		if st == nil || *st {
			return fmt.Errorf("automountServiceAccountToken must be false")
		}
		return nil
	})

	RunTest("secret access", func(j Job) error {
		secrets := sets.NewString()
		hasEntrypoint := false
		for _, c := range j.Base.Spec.Containers {
			if len(c.Command) > 0 && c.Command[0] == "entrypoint" {
				hasEntrypoint = true
			}
			for _, e := range c.Env {
				if e.Name == "GCP_SECRETS" {
					gcpSecrets := []Secret{}
					if err := json.Unmarshal([]byte(e.Value), &gcpSecrets); err != nil {
						return err
					}
					for _, s := range gcpSecrets {
						secrets.Insert(s.Project + "/" + s.Name)
					}
				}
			}
		}
		if secrets.Len() == 0 {
			return nil
		}

		if !hasEntrypoint {
			return fmt.Errorf("jobs with secrets must use entrypoint")
		}
		allowedSecret := strings.HasPrefix(j.Name, "release-notes") &&
			sets.NewString("istio-prow-build/github-read_github_read").IsSuperset(secrets)
		if !allowedSecret && j.Type == Presubmit {
			return fmt.Errorf("jobs with secrets %v cannot be presubmits", secrets.UnsortedList())
		}
		return nil
	})
}

func BuildRunTest(t *testing.T) func(name string, f func(j Job) error) {
	jobs := LoadJobs(t)
	return func(name string, f func(j Job) error) {
		t.Run(name, func(t *testing.T) {
			for _, j := range jobs {
				if err := f(j); err != nil {
					t.Errorf("job %v: %v", j.Name, err)
				}
			}
		})
	}
}

func LoadJobs(t *testing.T) []Job {
	const jobsPath = "../cluster/jobs"
	const configPath = "../config.yaml"
	c, err := config.LoadStrict(configPath, jobsPath, nil, "")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	jc := c.JobConfig
	var jobs []Job
	for repo, repoJobs := range jc.PresubmitsStatic {
		for _, job := range repoJobs {
			jobs = append(jobs, Job{
				Name:    job.Name,
				RepoOrg: repo,
				Type:    Presubmit,
				Base:    job.JobBase,
			})
		}
	}
	for repo, repoJobs := range jc.PostsubmitsStatic {
		for _, job := range repoJobs {
			jobs = append(jobs, Job{
				Name:    job.Name,
				RepoOrg: repo,
				Type:    Postsubmit,
				Base:    job.JobBase,
			})
		}
	}
	for _, job := range jc.Periodics {
		jobs = append(jobs, Job{
			Name: job.Name,
			Type: Periodic,
			Base: job.JobBase,
		})
	}
	return jobs
}

type JobType string

const (
	Presubmit  JobType = "presubmit"
	Postsubmit JobType = "postsubmit"
	Periodic   JobType = "periodic"
)

type Volumes = string

var AllVolumes = sets.NewString(
	GithubRelease,
	GithubTesting,
	GithubTestingSSH,
	BuildCache,
	Docker,
	Grafana,
	Netrc,
	SSHKey,
	IstioReleaseGCP,
	Cgroups,
	Modules,
)

var LowPrivilegeVolumes = sets.NewString(
	BuildCache,
	Cgroups,
	Modules,
)

var (
	PrivateVolumes = sets.NewString(Netrc, SSHKey)
	ReleaseVolumes = sets.NewString(GithubRelease, IstioReleaseGCP, Grafana, Docker)
)

const (
	GithubRelease    Volumes = "github-release"
	GithubTesting    Volumes = "github-testing"
	GithubTestingSSH Volumes = "github-testing-ssh"

	IstioReleaseGCP Volumes = "istio-release_gcp"

	Docker  Volumes = "docker"
	Grafana Volumes = "grafana"

	BuildCache Volumes = "buildcache"
	Cgroups    Volumes = "cgroups"
	Modules    Volumes = "modules"

	Netrc  Volumes = "netrc"
	SSHKey Volumes = "ssh-key"
)

type Job struct {
	Name    string
	RepoOrg string
	Type    JobType
	Base    config.JobBase
}

func (j Job) Org() string {
	org, _, _ := strings.Cut(j.RepoOrg, "/")
	return org
}

func (j Job) Repo() string {
	_, repo, _ := strings.Cut(j.RepoOrg, "/")
	return repo
}

func (j Job) BaseName() string {
	base, _, _ := strings.Cut(j.Name, "_")
	return base
}

func (j Job) Volumes() sets.String {
	r := sets.NewString()
	for _, v := range j.Base.Spec.Volumes {
		if v.Secret != nil {
			switch v.Secret.SecretName {
			case "oauth-token":
				r.Insert(GithubTesting)
			case "istio-testing-robot-ssh-key":
				r.Insert(GithubTestingSSH)
			case "rel-pipeline-service-account":
				r.Insert(IstioReleaseGCP)
			case "rel-pipeline-github":
				r.Insert(GithubRelease)
			case "grafana-token":
				r.Insert(Grafana)
			case "rel-pipeline-docker-config":
				r.Insert(Docker)
			case "netrc-secret":
				r.Insert(Netrc)
			case "ssh-key-secret":
				r.Insert(SSHKey)
			default:
				r.Insert("unknown secret/" + v.Secret.SecretName)
			}
		} else if v.HostPath != nil {
			switch v.HostPath.Path {
			case "/var/tmp/prow/cache":
				r.Insert(BuildCache)
			case "/sys/fs/cgroup":
				r.Insert(Cgroups)
			case "/lib/modules":
				r.Insert(Modules)
			default:
				r.Insert("unknown hostpath/" + v.HostPath.Path)
			}
		} else if v.EmptyDir != nil {
			// no issues here, just skip it
		} else {
			panic(fmt.Sprintf("unknown volume: %+v", v))
		}
	}
	return r
}

func (j Job) ServiceAccount() string {
	return j.Base.Spec.ServiceAccountName
}

type Sensitivity int

const (
	LowPrivilege Sensitivity = iota
	MediumPrivilege
	HighPrivilege
)

var ServiceAccounts = map[string]Sensitivity{
	"":                    LowPrivilege, // Default is prowjob-default-sa
	"prowjob-rbe":         MediumPrivilege,
	"prowjob-github-read": MediumPrivilege,
	"prowjob-default-sa":  LowPrivilege,
	"prow-deployer":       HighPrivilege,
	"testgrid-updater":    HighPrivilege,
	"prowjob-private-sa":  LowPrivilege,
	"prowjob-advanced-sa": HighPrivilege,
}

var PrivateServiceAccounts = sets.NewString(
	"prowjob-private-sa",
)

type Secret struct {
	Name    string `json:"secret,omitempty"`
	Project string `json:"project,omitempty"`
	Env     string `json:"env,omitempty"`
}

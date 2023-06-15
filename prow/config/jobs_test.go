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
		// legacy
		if strings.HasPrefix(j.Name, "release-notes") || strings.HasPrefix(j.Name, "update-ref-docs-dry-run") {
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
		releaseJob := j.Repo == "istio/release-builder" && j.Type == Postsubmit
		// TODO: these shouldn't need grafana or docker, and they actually don't - the private cluster has empty secrets
		privateReleaseJob := j.Repo == "istio-private/release-builder" && j.Type == Postsubmit
		baseImageBuilder := ((j.Repo == "istio/istio" && j.Type == Postsubmit) || j.Type == Periodic) && j.BaseName() == "build-base-images"
		if releaseJob || privateReleaseJob || baseImageBuilder {
			return nil
		}
		usesReleaseVolumes := j.Volumes().Intersection(ReleaseVolumes).Len() > 0
		if usesReleaseVolumes {
			return fmt.Errorf("only release jobs can use release volumes, found %v", j.Volumes().Intersection(ReleaseVolumes).UnsortedList())
		}
		return nil
	})

	// check to make sure we did not miss any volumes. This may just mean we need to update the test.
	RunTest("known service accounts only", func(j Job) error {
		if !AllServiceAccounts.Has(j.ServiceAccount()) {
			return fmt.Errorf("unknown service account: %q", j.ServiceAccount())
		}
		return nil
	})
	RunTest("presubmit jobs do not use privileged service accounts", func(j Job) error {
		if j.Type != Presubmit {
			return nil
		}
		// legacy
		if strings.Contains(j.Name, "_proxy") ||
			strings.HasPrefix(j.Name, "containers-test") ||
			strings.HasPrefix(j.Name, "dry-run_release-builder") ||
			strings.HasPrefix(j.Name, "release-test") ||
			j.Name == "benchmark-check_tools" {
			return nil
		}
		// Private volumes are handled in another test
		if !LowPrivServiceAccounts.Has(j.ServiceAccount()) {
			return fmt.Errorf("presubmit job using privileged service account: %q", j.ServiceAccount())
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
				Name: job.Name,
				Repo: repo,
				Type: Presubmit,
				Base: job.JobBase,
			})
		}
	}
	for repo, repoJobs := range jc.PostsubmitsStatic {
		for _, job := range repoJobs {
			jobs = append(jobs, Job{
				Name: job.Name,
				Repo: repo,
				Type: Postsubmit,
				Base: job.JobBase,
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
	Name string
	Repo string
	Type JobType
	Base config.JobBase
}

func (j Job) Org() string {
	org, _, _ := strings.Cut(j.Repo, "/")
	return org
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

var AllServiceAccounts = sets.NewString(
	"",
	"prow-deployer",
	"testgrid-updater",
	"gencred-refresher",
	"prowjob-private-sa",
	"prowjob-advanced-sa",
	"prowjob-default-sa",
)

var LowPrivServiceAccounts = sets.NewString(
	"", // Default is prowjob-default-sa
	"prowjob-default-sa",
)

var PrivateServiceAccounts = sets.NewString(
	"prowjob-private-sa",
)

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
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/prow/pkg/config"
	"sigs.k8s.io/yaml"
)

var (
	PrivateClusters = sets.NewString("private")
	PublicClusters  = sets.NewString("", "default", "prow-build", "test-infra-trusted")

	// ReadOnlySecrets are GCP secrets that grant read-only access to public
	// resources. They are safe to expose on presubmits and under any service
	// account, so the secret-related checks ignore them entirely.
	ReadOnlySecrets = sets.NewString(
		"istio-testing/cf_r2_public_buckets_ro_credentials",
	)
)

func TestJobs(t *testing.T) {
	RunTest := BuildRunTest(t)

	RunTest("tests use correct cluster", func(j Job) error {
		switch j.Org() {
		case "istio-private":
			if !PrivateClusters.Has(j.Base.Cluster) {
				return fmt.Errorf("private org must use private cluster, got %v", j.Base.Cluster)
			}
		case "istio", "istio-ecosystem":
			if !PublicClusters.Has(j.Base.Cluster) {
				return fmt.Errorf("primary org must use a public cluster, got: %v", j.Base.Cluster)
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
		if j.RepoOrg == "istio/test-infra" {
			// OK to run in trusted cluster
			return nil
		}
		if j.Type == Periodic {
			return nil
		}
		// Otherwise need allow-listed job only
		Allowed := sets.NewString(
			"sync-org_community_postsubmit",
			"deploy-policybot_bots_postsubmit",
		)
		if Allowed.Has(j.Name) {
			return nil
		}
		return fmt.Errorf("not allowed to run in trusted cluster")
	})

	RunTest("secure jobs do not use insecure caches", func(j Job) error {
		if j.Base.Cluster != "test-infra-trusted" {
			return nil
		}
		if j.Volumes().Has(BuildCache) {
			return fmt.Errorf("trusted jobs cannot use caches")
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
	RunTest("untrusted clusters do not use privileged volumes", func(j Job) error {
		if j.Base.Cluster == "test-infra-trusted" {
			return nil
		}
		priv := j.Volumes().Difference(LowPrivilegeVolumes).Difference(PrivateVolumes)
		if len(priv) == 0 {
			return nil
		}
		return fmt.Errorf("privileged volume must run in trusted cluster: %v", priv.UnsortedList())
	})
	RunTest("private volumes only used in private jobs", func(j Job) error {
		private := j.Org() == "istio-private"
		usesPrivate := j.Volumes().Intersection(PrivateVolumes).Len() > 0
		if usesPrivate && !private {
			return fmt.Errorf("only private jobs can use private volumes")
		}
		return nil
	})
	RunTest("org volumes only used in org jobs", func(j Job) error {
		orgJob := (j.RepoOrg == "istio/community" && j.Type == Postsubmit) ||
			(j.Name == "ci-test-infra-branchprotector" && j.Type == Periodic) ||
			// TODO: move these to use `github-istio-testing`
			(j.Name == "ci-prow-autobump" && j.Type == Periodic) ||
			(j.Name == "ci-prow-autobump-for-auto-deploy" && j.Type == Periodic)
		if orgJob {
			return nil
		}
		usesOrgVolume := j.Volumes().Has(GithubTestingOrgAdmin)
		if usesOrgVolume {
			return fmt.Errorf("only organization jobs can use organization volumes, found %v", j.Volumes())
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
			if j.Type == Presubmit {
				return fmt.Errorf("privileged service accounts cannot run as presubmit")
			}
			releaseJob := (j.RepoOrg == "istio/release-builder" && j.Type == Postsubmit) ||
				(strings.HasPrefix(j.Name, "build-base-images") && j.Type != Presubmit)
			if j.ServiceAccount() == "prowjob-release" && !releaseJob {
				return fmt.Errorf("only release jobs can use prowjob-release account")
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
		// Only 'prod' label is set on nodes in trusted cluster
		if j.Base.Cluster == "test-infra-trusted" {
			allowed := sets.NewString("prod", "kubernetes.io/arch")
			for k, v := range j.Base.Spec.NodeSelector {
				if !allowed.Has(k) {
					return fmt.Errorf("trusted cluster doesn't have nodes matching '%v=%v'", k, v)
				}
			}
			return nil
		}
		validSelectors := []map[string]string{}
		for _, arch := range []string{"amd64", "arm64"} {
			for _, tpe := range []string{"test-pool", "build-pool", "trusted"} {
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
					if e.Value == "" {
						continue
					}
					if err := json.Unmarshal([]byte(e.Value), &gcpSecrets); err != nil {
						return err
					}
					for _, s := range gcpSecrets {
						if ReadOnlySecrets.Has(s.Project + "/" + s.Name) {
							continue
						}
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
		if !allowedSecret && j.Type == Presubmit && !PrivateClusters.Has(j.Base.Cluster) {
			return fmt.Errorf("jobs with secrets %v cannot be presubmits", secrets.UnsortedList())
		}

		if secrets.Len() == 1 && secrets.Has("istio-testing/cf_r2_istio-prow_credentials") {
			// All pods already have access to this secret, as its needed to upload prowjob results to R2.
			return nil
		}
		secretSA := SecretServiceAccounts.Has(j.ServiceAccount())
		// Private cluster jobs run as prowjob-private by default (set via
		// cluster-wide default_service_account_name in prow/config.yaml,
		// not on the podSpec), so the SA appears empty here.
		if !secretSA && PrivateClusters.Has(j.Base.Cluster) && j.ServiceAccount() == "" {
			secretSA = true
		}
		if !secretSA {
			return fmt.Errorf("service account %v does not have Secrets access", j.ServiceAccount())
		}
		return nil
	})

	RunTest("trusted pool isolation", func(j Job) error {
		if j.Base.Cluster == "test-infra-trusted" {
			return nil
		}
		pool := j.Base.Spec.NodeSelector["testing"]
		if PrivateClusters.Has(j.Base.Cluster) && pool == "trusted" {
			return fmt.Errorf("private cluster does not have a trusted pool")
		}
		if ServiceAccounts[j.ServiceAccount()] == HighPrivilege && pool != "trusted" {
			return fmt.Errorf("high privilege service account %q must run on trusted pool", j.ServiceAccount())
		}
		if pool != "trusted" {
			return nil
		}
		if j.Type == Presubmit {
			return fmt.Errorf("presubmit jobs must not run on trusted pool")
		}
		if !hasTrustedToleration(j.Base.Spec.Tolerations) {
			return fmt.Errorf("job does not tolerate trusted pool taint")
		}
		return nil
	})
	RunTest("bazel cache isolation", func(j Job) error {
		if j.Base.Spec == nil || len(j.Base.Spec.Containers) == 0 {
			return nil
		}
		cacheURL := ""
		for _, env := range j.Base.Spec.Containers[0].Env {
			if env.Name == "BAZEL_BUILD_RBE_CACHE" {
				cacheURL = env.Value
			}
		}
		var cacheSidecar *v1.Container
		for i := range j.Base.Spec.InitContainers {
			if j.Base.Spec.InitContainers[i].Name == "bazel-remote" {
				cacheSidecar = &j.Base.Spec.InitContainers[i]
				break
			}
		}
		if cacheURL == "" && cacheSidecar == nil {
			return nil
		}
		if cacheURL != "http://localhost:8080" {
			return fmt.Errorf("bazel cache must use localhost, got %q", cacheURL)
		}
		if cacheSidecar == nil {
			return fmt.Errorf("bazel cache URL configured without sidecar")
		}
		if !sets.New(cacheSidecar.Args...).Has("--s3.bucket=$(BAZEL_REMOTE_BUCKET)") {
			return fmt.Errorf("bazel cache sidecar must use the configured bucket")
		}
		expectedBucket := "istio-prow-bazel-cache"
		expectedServiceAccount := "prowjob-proxy-presubmit"
		if j.Org() == "istio-private" {
			expectedBucket = "istio-prow-bazel-cache-private"
			expectedServiceAccount = "prowjob-private"
		} else if j.Type == Postsubmit {
			expectedBucket = "istio-prow-bazel-cache-postsubmit"
			expectedServiceAccount = "prowjob-proxy-postsubmit"
		}
		if j.Type != Presubmit && j.Type != Postsubmit {
			return fmt.Errorf("bazel cache sidecar is not allowed for %v jobs", j.Type)
		}
		cacheBucket := ""
		for _, env := range cacheSidecar.Env {
			if env.Name == "BAZEL_REMOTE_BUCKET" {
				cacheBucket = env.Value
			}
		}
		if cacheBucket != expectedBucket {
			return fmt.Errorf("bazel cache sidecar must use bucket %q", expectedBucket)
		}
		if j.ServiceAccount() != expectedServiceAccount {
			return fmt.Errorf("bazel cache job must use service account %q, got %q", expectedServiceAccount, j.ServiceAccount())
		}
		return nil
	})
}

func TestSensitiveRequirementsIncludeTrusted(t *testing.T) {
	files, err := filepath.Glob("jobs/*.yaml")
	if err != nil {
		t.Fatalf("failed to find source jobs: %v", err)
	}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), ".") {
			continue
		}
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read %s: %v", file, err)
		}
		source := struct {
			Jobs []struct {
				Name               string   `json:"name"`
				Cluster            string   `json:"cluster"`
				Requirements       []string `json:"requirements"`
				ServiceAccountName string   `json:"service_account_name"`
			} `json:"jobs"`
		}{}
		if err := yaml.Unmarshal(content, &source); err != nil {
			t.Fatalf("failed to parse %s: %v", file, err)
		}
		for _, job := range source.Jobs {
			if job.Cluster == "test-infra-trusted" {
				continue
			}
			requirements := sets.New(job.Requirements...)
			requiresTrusted := ServiceAccounts[job.ServiceAccountName] == HighPrivilege
			for requirement := range requirements {
				if isWriteR2Requirement(requirement) {
					requiresTrusted = true
				}
			}
			if requiresTrusted && !requirements.Has("trusted") {
				t.Errorf("%s: job %q must explicitly require trusted", file, job.Name)
			}
		}
	}
}

func isWriteR2Requirement(requirement string) bool {
	return strings.HasPrefix(requirement, "cf-r2-istio-") &&
		requirement != "cf-r2-public-ro-auth" &&
		!strings.HasSuffix(requirement, "-private-auth")
}

func hasTrustedToleration(tolerations []v1.Toleration) bool {
	for _, toleration := range tolerations {
		if toleration.Key == "testing" &&
			toleration.Operator == v1.TolerationOpEqual &&
			toleration.Value == "trusted" &&
			toleration.Effect == v1.TaintEffectNoSchedule {
			return true
		}
	}
	return false
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
	var files []string
	for _, org := range []string{"istio", "istio-private"} {
		orgFiles, err := filepath.Glob(filepath.Join("../cluster/jobs", org, "*/*.gen.yaml"))
		if err != nil {
			t.Fatalf("failed to find generated jobs: %v", err)
		}
		files = append(files, orgFiles...)
	}
	var jobs []Job
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read %s: %v", file, err)
		}
		jc := config.JobConfig{}
		if err := yaml.UnmarshalStrict(content, &jc); err != nil {
			t.Fatalf("failed to parse %s: %v", file, err)
		}
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

var AllVolumes = sets.New(
	GithubTestingOrgAdmin,
	GithubTestingPusher,
	GithubTestingSSH,
	BuildCache,
	Netrc,
	SSHKey,
	Cgroups,
	Modules,
)

var LowPrivilegeVolumes = sets.New(
	BuildCache,
	Cgroups,
	Modules,
)

var PrivateVolumes = sets.New(Netrc, SSHKey)

const (
	GithubTestingOrgAdmin Volumes = "github-testing"
	GithubTestingPusher   Volumes = "github-testing-pusher"
	GithubTestingSSH      Volumes = "github-testing-ssh"

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

func (j Job) Volumes() sets.Set[string] {
	r := sets.New[string]()
	for _, v := range j.Base.Spec.Volumes {
		if v.Secret != nil {
			switch v.Secret.SecretName {
			case "oauth-token":
				r.Insert(GithubTestingOrgAdmin)
			case "github-istio-testing-pusher":
				r.Insert(GithubTestingOrgAdmin)
			case "istio-testing-robot-ssh-key":
				r.Insert(GithubTestingSSH)
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
	"":                             LowPrivilege, // Default is prowjob-default-sa
	"prowjob-default-sa":           LowPrivilege,
	"prowjob-proxy-presubmit":      LowPrivilege,
	"prowjob-private":              LowPrivilege,
	"prowjob-rbe":                  MediumPrivilege,
	"prowjob-github-read":          MediumPrivilege,
	"prow-deployer":                HighPrivilege,
	"testgrid-updater":             HighPrivilege,
	"prowjob-testing-write":        HighPrivilege,
	"prowjob-proxy-postsubmit":     HighPrivilege,
	"prowjob-github-istio-testing": HighPrivilege,
	"prowjob-release":              HighPrivilege,
	"prowjob-build-tools":          HighPrivilege,
	"prowjob-bots-deployer":        HighPrivilege,
}

var PrivateServiceAccounts = sets.NewString(
	"prowjob-private",
)

// SA with Secret access
var SecretServiceAccounts = sets.NewString(
	"prowjob-github-istio-testing",
	"prowjob-github-read",
	"prowjob-release",
	"prowjob-build-tools",
	"prowjob-testing-write",
	"prowjob-proxy-postsubmit",
	"prowjob-private",
)

type Secret struct {
	Name    string `json:"secret,omitempty"`
	Project string `json:"project,omitempty"`
	Env     string `json:"env,omitempty"`
	File    string `json:"file,omitempty"`
}

// Copyright 2018 Istio Authors
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
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/test-infra/prow/config"
	_ "k8s.io/test-infra/prow/hook"
)

var (
	configPath    = flag.String("config", "../config.yaml", "Path to prow config")
	jobConfigPath = flag.String("job-config", "../cluster/jobs/", "Path to prow job config")
)

// Loaded at TestMain.
var c *config.Config

func TestMain(m *testing.M) {
	flag.Parse()

	cfg, err := config.Load(*configPath, *jobConfigPath, []string{}, "")
	if err != nil {
		fmt.Printf("Could not load config: %v\n", err)
		os.Exit(1)
	}
	c = cfg

	os.Exit(m.Run())
}

func TestConfig(t *testing.T) {
	one := 1
	two := 2
	yes := true
	var no bool
	cases := []struct {
		name               string
		org                string
		repo               string
		branch             string
		unprotected        bool
		approvers          *int
		codeOwners         *bool
		expectedContexts   []string
		unexpectedContexts []string
		teams              []string
		anyoneCanMerge     *bool
	}{
		{
			name:   "protect istio master",
			org:    "istio",
			repo:   "istio",
			branch: "master",
			expectedContexts: []string{
				"cla/google",
			},
		},
		{
			name:       "api requires code owner reviews",
			org:        "istio",
			repo:       "api",
			branch:     "master",
			codeOwners: &yes,
		},
		{
			name:       "test-infra requires code owners",
			org:        "istio",
			repo:       "test-infra",
			branch:     "master",
			codeOwners: &yes,
		},
		{
			name:      "api requires 2 approvers",
			org:       "istio",
			repo:      "api",
			branch:    "master",
			approvers: &two,
		},
		{
			name:      "test-infra requires 1 approving review",
			org:       "istio",
			repo:      "test-infra",
			branch:    "master",
			approvers: &one,
		},
		{
			name:       "operator requires 1 approving review and code owners",
			org:        "istio",
			repo:       "operator",
			branch:     "master",
			approvers:  &one,
			codeOwners: &yes,
		},
		{
			name:        "istio unprotected by default",
			org:         "istio",
			repo:        "istio",
			branch:      "something-random",
			unprotected: true,
		},
		{
			name:   "api release 1.0 requires admin merge",
			org:    "istio",
			repo:   "api",
			branch: "release-1.0",
			expectedContexts: []string{
				"merges-blocked-needs-admin",
			},
		},
		{
			name:   "istio 1.2 requires circleci and admin merges",
			org:    "istio",
			repo:   "istio",
			branch: "release-1.2",
			expectedContexts: []string{
				"ci/circleci: build",
				"merges-blocked-needs-admin",
			},
		},
		{
			name:   "istio 1.3 protected",
			org:    "istio",
			repo:   "istio",
			branch: "release-1.3",
		},
		{
			name:   "istio 1.4 protected",
			org:    "istio",
			repo:   "istio",
			branch: "release-1.4",
		},
		{
			name:               "test-infra not blocked by admins or circleci",
			org:                "istio",
			repo:               "test-infra",
			branch:             "master",
			expectedContexts:   []string{"cla/google"},
			unexpectedContexts: []string{"merges-blocked-need-admin"},
		},
		{
			name:   "operator protects master",
			org:    "istio",
			repo:   "operator",
			branch: "master",
		},
		{
			name:   "operator protects release-0.1",
			org:    "istio",
			repo:   "operator",
			branch: "release-0.1",
		},
		{
			name:   "operator protects release-1.1",
			org:    "istio",
			repo:   "operator",
			branch: "release-1.1",
		},
		{
			name:        "all istio repos define a policy",
			org:         "istio",
			repo:        "random-repo",
			branch:      "master",
			unprotected: true,
		},
		{
			name:        "all istio-ecosystem repos define a policy",
			org:         "istio-ecosystem",
			repo:        "random-repo",
			branch:      "master",
			unprotected: true,
		},
		{
			name:   "istio-ecosystem/authservice protects master",
			org:    "istio-ecosystem",
			repo:   "authservice",
			branch: "master",
		},
		{
			name:   "release-1.1 team can merge into api",
			org:    "istio",
			repo:   "api",
			branch: "release-1.1",
			teams:  []string{"release-managers-1-1"},
		},
		{
			name:   "release-1.1 team can merge into istio",
			org:    "istio",
			repo:   "istio",
			branch: "release-1.1",
			teams:  []string{"release-managers-1-1"},
		},
		{
			name:   "release-1.2 team can merge into api",
			org:    "istio",
			repo:   "api",
			branch: "release-1.2",
			teams:  []string{"release-managers-1-2"},
		},
		{
			name:   "release-1.3 team can merge into api",
			org:    "istio",
			repo:   "api",
			branch: "release-1.3",
			teams:  []string{"release-managers-1-3"},
		},
		{
			name:   "release-1.2 team can merge into istio",
			org:    "istio",
			repo:   "istio",
			branch: "release-1.2",
			teams:  []string{"release-managers-1-2", "repo-admins"},
		},
		{
			name:   "release-1.3 team can merge into istio",
			org:    "istio",
			repo:   "istio",
			branch: "release-1.3",
			teams:  []string{"release-managers-1-3", "repo-admins"},
		},
		{
			name:   "release-1.4 team can merge into istio",
			org:    "istio",
			repo:   "istio",
			branch: "release-1.4",
			teams:  []string{"release-managers-1-4", "repo-admins"},
		},
		{
			name:           "master mergify branches allow anyone to merge",
			org:            "istio",
			repo:           "api",
			branch:         "master",
			anyoneCanMerge: &no,
		},
		{
			name:   "test-infra master restricts merges to admins",
			org:    "istio",
			repo:   "test-infra",
			branch: "master",
			teams:  []string{"repo-admins"},
		},
		{
			name:   "istio master restricts merges to admins",
			org:    "istio",
			repo:   "istio",
			branch: "master",
			teams:  []string{"repo-admins"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bp, err := c.GetBranchProtection(tc.org, tc.repo, tc.branch, nil)
			switch {
			case err != nil:
				t.Errorf("call failed: %v", err)
				return
			case bp == nil:
				t.Error("undefined policy")
				return
			case bp.Protect == nil:
				t.Error("undefined protection")
				return
			case *bp.Protect == tc.unprotected:
				t.Errorf("protect: %t (actual) but we expect %t", *bp.Protect, !tc.unprotected)
			}

			if bp.RequiredStatusChecks == nil {
				bp.RequiredStatusChecks = &config.ContextPolicy{}
			}
			if bp.Restrictions == nil {
				bp.Restrictions = &config.Restrictions{}
			}
			if bp.RequiredPullRequestReviews == nil {
				bp.RequiredPullRequestReviews = &config.ReviewPolicy{}
			}

			actual := sets.NewString(bp.RequiredStatusChecks.Contexts...)
			if missing := sets.NewString(tc.expectedContexts...).Difference(actual); len(missing) > 0 {
				t.Errorf("missing contexts: %v", missing.List())
			}
			if unexpected := sets.NewString(tc.unexpectedContexts...).Intersection(actual); len(unexpected) > 0 {
				t.Errorf("unexpected contexts: %v", unexpected.List())
			}
			if tc.anyoneCanMerge != nil && *tc.anyoneCanMerge && bp.Restrictions.Users != nil {
				t.Errorf("should allow anyone to merge, but restricting to %v", bp.Restrictions.Users)
			}
			if tc.anyoneCanMerge != nil && *tc.anyoneCanMerge && bp.Restrictions.Teams != nil {
				t.Errorf("should allow anyone to merge, but restricting to %v", bp.Restrictions.Teams)
			}
			actual = sets.NewString(bp.Restrictions.Teams...)
			if missing := sets.NewString(tc.teams...).Difference(actual); len(missing) > 0 {
				t.Errorf("missing teams: %v", missing.List())
			}
			if tc.approvers != nil && *tc.approvers != *bp.RequiredPullRequestReviews.Approvals {
				t.Errorf("%d actual approvers != expected %d", *bp.RequiredPullRequestReviews.Approvals, *tc.approvers)
			}
			if tc.codeOwners != nil && *tc.codeOwners != *bp.RequiredPullRequestReviews.RequireOwners {
				t.Errorf("%t actual codeOwners != expected %t", *bp.RequiredPullRequestReviews.RequireOwners, *tc.codeOwners)
			}
		})
	}
}

func TestTrustedJobs(t *testing.T) {
	const trusted = "test-infra-trusted"
	trustedPath := path.Join(*jobConfigPath, "istio", "test-infra")

	// Presubmits may not use trusted clusters.
	for _, pre := range c.AllStaticPresubmits(nil) {
		if pre.Cluster == trusted {
			t.Errorf("%s: presubmits cannot use trusted clusters", pre.Name)
		}
	}

	// Trusted postsubmits must be defined in trustedPath
	for _, post := range c.AllStaticPostsubmits(nil) {
		if post.Cluster != trusted {
			continue
		}
		if !strings.HasPrefix(post.SourcePath, trustedPath) {
			t.Errorf("%s defined in %s may not run in trusted cluster", post.Name, post.SourcePath)
		}
	}

	// Trusted periodics must be defined in trustedPath
	for _, per := range c.AllPeriodics() {
		if per.Cluster != trusted {
			continue
		}
		if !strings.HasPrefix(per.SourcePath, trustedPath) {
			t.Errorf("%s defined in %s may not run in trusted cluster", per.Name, per.SourcePath)
		}
	}
}

func TestPresets(t *testing.T) {
	known := sets.NewString()
	for _, p := range c.Presets {
		if len(p.Labels) != 1 {
			t.Fatalf("Istio presets are expected to have a single label")
		}
		for k := range p.Labels {
			known.Insert(k)
			if !strings.HasPrefix(k, "preset-") {
				t.Fatalf("preset must start with 'preset-': %v", k)
			}
		}
	}
	unused := sets.NewString(known.UnsortedList()...)
	for _, pre := range c.AllStaticPresubmits(nil) {
		for k := range pre.Labels {
			if strings.HasPrefix(k, "preset-") {
				unused.Delete(k)
				if !known.Has(k) {
					t.Fatalf("%v: unknown preset %v", pre.Name, k)
				}
			}
		}
	}
	for _, post := range c.AllStaticPostsubmits(nil) {
		for k := range post.Labels {
			if strings.HasPrefix(k, "preset-") {
				unused.Delete(k)
				if !known.Has(k) {
					t.Fatalf("%v: unknown preset %v", post.Name, k)
				}
			}
		}
	}

	for _, per := range c.AllPeriodics() {
		for k := range per.Labels {
			if strings.HasPrefix(k, "preset-") {
				unused.Delete(k)
				if !known.Has(k) {
					t.Fatalf("%v: unknown preset %v", per.Name, k)
				}
			}
		}
	}

	if len(unused) > 0 {
		t.Fatalf("unused presets %v should be removed", unused.List())
	}
}

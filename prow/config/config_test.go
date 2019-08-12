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
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/test-infra/prow/config"
	_ "k8s.io/test-infra/prow/hook"
	"k8s.io/test-infra/prow/plugins"
)

func TestConfig(t *testing.T) {
	cfg, err := config.Load("../config.yaml", "../cluster/jobs/")
	if err != nil {
		t.Fatalf("could not read configs: %v", err)
	}
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
			name:   "collab-galley restricted to hackers",
			org:    "istio",
			repo:   "istio",
			branch: "collab-galley",
			teams:  []string{"istio-hackers"},
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
			name:        "all istio-releases repos define a policy",
			org:         "istio-releases",
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
			name:   "istio-releases/pipeline protects master",
			org:    "istio-releases",
			repo:   "pipeline",
			branch: "master",
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
			name:   "release-1.2 team can merge into istio",
			org:    "istio",
			repo:   "istio",
			branch: "release-1.2",
			teams:  []string{"release-managers-1-2", "repo-admins"},
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
			bp, err := cfg.GetBranchProtection(tc.org, tc.repo, tc.branch)
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

// Make sure that our plugins are valid.
func TestPlugins(t *testing.T) {
	pa := &plugins.ConfigAgent{}
	if err := pa.Load("../plugins.yaml"); err != nil {
		t.Fatalf("could not load plugins: %v", err)
	}
}

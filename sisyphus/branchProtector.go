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

package sisyphus

import (
	u "istio.io/test-infra/toolbox/util"
)

type branchProtector struct {
	githubClnt *u.GithubClient
	repo       string
	branch     string
}

func newBranchProtector(owner, token, repo, branch string) *branchProtector {
	return &branchProtector{
		githubClnt: u.NewGithubClient(owner, token),
		repo:       repo,
		branch:     branch,
	}
}

func (p *branchProtector) process(failures []failure) {
	if failures != nil {
		u.BlockMergingOnBranch(p.githubClnt, p.repo, p.branch)
	} else {
		u.UnBlockMergingOnBranch(p.githubClnt, p.repo, p.branch)
	}
}

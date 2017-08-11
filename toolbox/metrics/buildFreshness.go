// Copyright 2017 Istio Authors
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
	"strings"
	"sync"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	u "istio.io/test-infra/toolbox/util"
)

const (
	istioDepsFile = "istio.deps"
)

// DepFreshness stores how many days behind the dependency SHA used by the last
// stable build is compared to the HEAD of the production branch of that dependency
type DepFreshness struct {
	Dep       u.Dependency
	Freshness time.Duration
}

func getBranchHeadTime(gclient *u.GithubClient, repo, branch string) (*time.Time, error) {
	sha, err := gclient.GetHeadCommitSHA(repo, branch)
	if err != nil {
		return nil, err
	}
	t, err := gclient.GetSHATime(repo, sha)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func getStableBuildFreshness(owner, repo, branch string) ([]DepFreshness, error) {
	githubClnt := u.NewGithubClientNoAuth(owner)
	var stats []DepFreshness
	pickledDeps, err := githubClnt.GetFileContent(repo, branch, istioDepsFile)
	if err != nil {
		return stats, err
	}
	deps, err := u.DeserializeDepsFromString(pickledDeps)
	if err != nil {
		return stats, err
	}
	var wg sync.WaitGroup
	var mutex = &sync.Mutex{} // used to synchronize access to stats and multiErr
	var multiErr error
	for _, dep := range deps {
		wg.Add(1)
		go func(dep u.Dependency) {
			latestTime, err := getBranchHeadTime(githubClnt, dep.RepoName, dep.ProdBranch)
			if err != nil {
				// multierror not thread safe
				mutex.Lock()
				multiErr = multierror.Append(multiErr, err)
				mutex.Unlock()
				return
			}
			var stableTime *time.Time
			if strings.Contains(dep.LastStableSHA, "-") || strings.Contains(dep.LastStableSHA, ".") {
				// the reference is a tag
				stableTime, err = githubClnt.GetTagPublishTime(dep.RepoName, dep.LastStableSHA)
			} else {
				stableTime, err = githubClnt.GetSHATime(dep.RepoName, dep.LastStableSHA)
			}
			if err != nil {
				mutex.Lock()
				multiErr = multierror.Append(multiErr, err)
				mutex.Unlock()
				return
			}
			lag := latestTime.Sub(*stableTime)
			mutex.Lock()
			stats = append(stats, DepFreshness{dep, lag})
			mutex.Unlock()
			wg.Done()
		}(dep)
	}
	wg.Wait()
	return stats, multiErr
}

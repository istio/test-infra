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
	"fmt"
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

func getBranchHeadTime(ghClient *u.GithubClient, repo, branch string) (*time.Time, error) {
	sha, err := ghClient.GetHeadCommitSHA(repo, branch)
	if err != nil {
		return nil, err
	}
	t, err := ghClient.GetCommitCreationTime(repo, sha)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func getCommitCreationTimeByRef(ghClient *u.GithubClient, repo, ref string) (*time.Time, error) {
	if u.SHARegex.MatchString(ref) {
		return ghClient.GetCommitCreationTime(repo, ref)
	} else if u.ReleaseTagRegex.MatchString(ref) {
		return ghClient.GetCommitCreationTimeByTag(repo, ref)
	}
	err := fmt.Errorf(
		"reference must be a SHA or release tag to get creation time, but was instead %s", ref)
	return nil, err
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
	multiErrAppend := func(err error) {
		// multierror not thread safe
		mutex.Lock()
		multiErr = multierror.Append(multiErr, err)
		mutex.Unlock()
	}
	for _, dep := range deps {
		wg.Add(1)
		go func(dep u.Dependency) {
			defer wg.Done()
			stableTime, err := getCommitCreationTimeByRef(githubClnt, dep.RepoName, dep.LastStableSHA)
			if err != nil {
				e := fmt.Errorf(
					"failed to get the committed time of the stable dependency named %s: %v",
					dep.Name, err)
				multiErrAppend(e)
				return
			}
			latestTime, err := getBranchHeadTime(githubClnt, dep.RepoName, dep.ProdBranch)
			if err != nil {
				e := fmt.Errorf(
					"failed to get the committed time of HEAD of branch %s on repo %s: %v",
					dep.ProdBranch, dep.RepoName, err)
				multiErrAppend(e)
				return
			}
			lag := latestTime.Sub(*stableTime)
			mutex.Lock()
			defer mutex.Unlock()
			stats = append(stats, DepFreshness{dep, lag})
		}(dep)
	}
	wg.Wait()
	return stats, multiErr
}

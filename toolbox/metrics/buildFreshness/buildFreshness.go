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

package buildFreshness

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

// DepFreshness stores how long the dependency SHA used by the last
// stable build is behind the HEAD of the production branch of that dependency
type DepFreshness struct {
	Dep u.Dependency
	Age time.Duration
}

func getBranchHeadTime(githubClnt *u.GithubClient, repo, branch string) (time.Time, error) {
	sha, err := githubClnt.GetHeadCommitSHA(repo, branch)
	if err != nil {
		return time.Time{}, err
	}
	t, err := githubClnt.GetCommitCreationTime(repo, sha)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func getCommitCreationTimeByRef(githubClnt *u.GithubClient, repo, ref string) (time.Time, error) {
	if u.SHARegex.MatchString(ref) {
		return githubClnt.GetCommitCreationTime(repo, ref)
	} else if u.ReleaseTagRegex.MatchString(ref) {
		return githubClnt.GetCommitCreationTimeByTag(repo, ref)
	}
	err := fmt.Errorf(
		"reference must be a SHA or release tag to get creation time, but was instead %s", ref)
	return time.Time{}, err
}

func getAgeMetric(githubClnt *u.GithubClient, dep *u.Dependency) (*DepFreshness, error) {
	stableTime, err := getCommitCreationTimeByRef(githubClnt, dep.RepoName, dep.LastStableSHA)
	if err != nil {
		e := fmt.Errorf(
			"failed to get the committed time of the stable dependency named %s: %v",
			dep.Name, err)
		return nil, e
	}
	latestTime, err := getBranchHeadTime(githubClnt, dep.RepoName, dep.ProdBranch)
	if err != nil {
		e := fmt.Errorf(
			"failed to get the committed time of HEAD of branch %s on repo %s: %v",
			dep.ProdBranch, dep.RepoName, err)
		return nil, e
	}
	lag := latestTime.Sub(stableTime)
	return &DepFreshness{*dep, lag}, nil
}

// GetAgeMetrics gives each dependency ages.
func GetAgeMetrics(owner, repo, branch string) ([]DepFreshness, error) {
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
			defer wg.Done()
			ageMetric, err := getAgeMetric(githubClnt, &dep)
			mutex.Lock()
			defer mutex.Unlock()
			if err != nil {
				multiErr = multierror.Append(multiErr, err)
				return
			}
			stats = append(stats, *ageMetric)
		}(dep)
	}
	wg.Wait()
	return stats, multiErr
}

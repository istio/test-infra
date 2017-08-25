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

package util

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
)

var (
	ci = NewCIState()
	// SHARegex matches commit SHA's
	SHARegex = regexp.MustCompile("^[a-z0-9]{40}$")
	// ReleaseTagRegex matches release tags
	ReleaseTagRegex = regexp.MustCompile("^[0-9]+.[0-9]+.[0-9]+$")
)

// CIState defines constants representing possible states of
// continuous integration tests
type CIState struct {
	Success string
	Failure string
	Pending string
	Error   string
}

// NewCIState creates a new CIState
func NewCIState() *CIState {
	return &CIState{
		Success: "success",
		Failure: "failure",
		Pending: "pending",
		Error:   "error",
	}
}

// GetRequiredCheckStatuses does NOT trust the given combined output but instead walk
// through the CI results, count states, and determine the final state
// as either pending, failure, or success
func GetRequiredCheckStatuses(combinedStatus *github.CombinedStatus,
	requiredStatusCheckContexts []string, skipContext func(string) bool) string {
	var failures, pending, successes int
	for _, status := range combinedStatus.Statuses {
		if !ContainsString(requiredStatusCheckContexts, *status.Context) {
			// This status check is not required for merging
			continue
		}
		if *status.State == ci.Error || *status.State == ci.Failure {
			if skipContext != nil && skipContext(*status.Context) {
				continue
			}
			failures++
		} else if *status.State == ci.Pending {
			pending++
		} else if *status.State == ci.Success {
			successes++
		} else {
			log.Printf("Check Status %s is unknown", *status.State)
		}
	}
	if pending > 0 {
		return ci.Pending
	} else if failures > 0 {
		return ci.Failure
	} else {
		return ci.Success
	}
}

// GetAPITokenFromFile returns the github api token from tokenFile
func GetAPITokenFromFile(tokenFile string) (string, error) {
	b, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(string(b[:]))
	return token, nil
}

// CloneRepoCheckoutBranch removes previous repo, clone to local machine,
// change directory into the repo, and checkout the given branch.
// Returns the absolute path to repo root
func CloneRepoCheckoutBranch(gclient *GithubClient, repo, baseBranch, newBranch string) (string, error) {
	if err := os.RemoveAll(repo); err != nil {
		return "", err
	}
	if _, err := ShellSilent(
		"git clone " + gclient.Remote(repo)); err != nil {
		return "", err
	}
	if err := os.Chdir(repo); err != nil {
		return "", err
	}
	if _, err := Shell("git checkout " + baseBranch); err != nil {
		return "", err
	}
	if newBranch != "" {
		if _, err := Shell("git checkout -b " + newBranch); err != nil {
			return "", err
		}
	}
	return os.Getwd()
}

// RemoveLocalRepo deletes the local git repo just cloned
func RemoveLocalRepo(absolutePathToRepo string) error {
	return os.RemoveAll(absolutePathToRepo)
}

// CreateCommitPushToRemote stages call local changes, create a commit,
// and push to remote tracking branch
func CreateCommitPushToRemote(branch, commitMsg string) error {
	if _, err := Shell("git commit -am " + commitMsg); err != nil {
		return err
	}
	_, err := Shell("git push --set-upstream origin " + branch)
	return err
}

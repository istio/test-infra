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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"istio.io/test-infra/toolbox/util"
)

var (
	repo       = flag.String("repo", "", "Update dependencies of only this repository")
	owner      = flag.String("owner", "istio", "Github Owner or org.")
	tokenFile  = flag.String("token_file", "", "File containing Github API Access Token.")
	githubClnt *githubClient
)

const (
	istioDepsFile = "istio.deps"
)

// Update the commit SHA reference in a given line from dependency file
// to the latest stable version
// returns the updated line
func replaceCommit(line string, dep dependency) (string, error) {
	commitSHA, err := githubClnt.getHeadCommitSHA(dep.RepoName, dep.ProdBranch)
	if err != nil {
		return line, err
	}
	idx := strings.Index(line, "\"")
	return line[:idx] + "\"" + commitSHA + "\",", nil
}

// Generates an MD5 digest of the version set of the repo dependencies
// useful in avoiding making duplicate branches of the same code change
func fingerPrint(repo string, deps []dependency) (string, error) {
	digest, err := githubClnt.getHeadCommitSHA(repo, "master")
	if err != nil {
		return "", err
	}
	for _, dep := range deps {
		commitSHA, err := githubClnt.getHeadCommitSHA(dep.RepoName, dep.ProdBranch)
		if err != nil {
			return "", err
		}
		digest = digest + commitSHA
	}
	return util.GetMD5Hash(digest), nil
}

// Update the commit SHA reference in the dependency file of dep
func updateDepFile(dep dependency) error {
	input, err := ioutil.ReadFile(dep.File)
	if err != nil {
		return err
	}
	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.Contains(line, dep.Name+" = ") {
			if lines[i], err = replaceCommit(line, dep); err != nil {
				return err
			}
		}
	}
	output := strings.Join(lines, "\n")
	return ioutil.WriteFile(dep.File, []byte(output), 0600)
}

// Assumes at the root of istio directory
// Runs the updateVersion.sh script
func updateIstioDeps() error {
	pilotSHA, err := githubClnt.getHeadCommitSHA("pilot", "stable")
	if err != nil {
		return err
	}
	mixerSHA, err := githubClnt.getHeadCommitSHA("mixer", "stable")
	if err != nil {
		return err
	}
	istioctlURL := fmt.Sprintf(
		"https://storage.googleapis.com/istio-artifacts/pilot/%s/artifacts/istioctl", pilotSHA)
	hub := "gcr.io/istio-testing"
	cmd := fmt.Sprintf("./install/updateVersion.sh -p %s,%s -x %s,%s -i %s",
		hub, pilotSHA, hub, mixerSHA, istioctlURL)
	_, err = util.Shell(cmd)
	return err
}

// Delete the local git repo just cloned
func cleanUp(repo string) error {
	if err := os.Chdir(".."); err != nil {
		return err
	}
	return os.RemoveAll(repo)
}

// Update the given repository so that it uses the latest dependency references
// push new branch to remote, create pull request on master,
// which is auto-merged after presumbit
func updateDeps(repo string) error {
	if err := os.RemoveAll(repo); err != nil {
		return err
	}
	if _, err := util.Shell("git clone " + githubClnt.remote(repo)); err != nil {
		return err
	}
	if err := os.Chdir(repo); err != nil {
		return err
	}
	defer func() {
		if err := cleanUp(repo); err != nil {
			log.Fatalf("Error during clean up: %v\n", err)
		}
	}()
	deps, err := getDeps(istioDepsFile)
	if err != nil {
		return err
	}
	depVersions, err := fingerPrint(repo, deps)
	if err != nil {
		return err
	}
	branch := "autoUpdateDeps" + depVersions
	exists, err := githubClnt.existBranch(repo, branch)
	if err != nil {
		return err
	}
	// if branch exists, stop here and do not create another PR of identical delta
	if err = githubClnt.closeFailedPullRequests(repo); exists || err != nil {
		return err
	}
	// TODO (chx) refactor github helper
	// TODO (chx) metrics
	// TODO (chx) updateIstioDeps() should also take care of istio-ca (auth)
	// TODO (chx) comment functions
	if _, err := util.Shell("git checkout -b " + branch); err != nil {
		return err
	}
	if repo == "istio" {
		if err := updateIstioDeps(); err != nil {
			return err
		}
	} else {
		for _, dep := range deps {
			if err := updateDepFile(dep); err != nil {
				return err
			}
		}
	}
	if _, err := util.Shell("git add *"); err != nil {
		return err
	}
	if _, err := util.Shell("git commit -m Update_Dependencies"); err != nil {
		return err
	}
	if _, err := util.Shell("git push --set-upstream origin " + branch); err != nil {
		return err
	}
	if err := githubClnt.createPullRequest(branch, repo); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	if *tokenFile == "" {
		log.Panicf("token_file not provided\n")
		return
	}
	token, err := util.GetAPITokenFromFile(*tokenFile)
	if err != nil {
		log.Panicf("Error accessing user supplied token_file: %v\n", err)
	}
	githubClnt, err = newGithubClient(*owner, token)
	if err != nil {
		log.Panicf("Error when initializing github client: %v\n", err)
	}
	if *repo != "" { // only update dependencies of this repo
		if err := updateDeps(*repo); err != nil {
			log.Panicf("Failed to udpate dependency: %v\n", err)
		}
	} else { // update dependencies of all repos in the istio project
		repos, err := githubClnt.listRepos()
		if err != nil {
			log.Panicf("Error when fetching list of repos: %v\n", err)
			return
		}
		for _, r := range repos {
			if err := updateDeps(r); err != nil {
				log.Panicf("Failed to udpate dependency: %v\n", err)
			}
		}
	}
}

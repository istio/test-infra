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
	repo       = flag.String("repo", "", "Optional. Update dependencies of only this repository")
	owner      = flag.String("owner", "istio", "Github Owner or org")
	tokenFile  = flag.String("token_file", "", "File containing Github API Access Token")
	baseBranch = flag.String("base_branch", "master", "Branch from which the deps update commit is based")
	hub        = flag.String("hub", "", "Where the testing images are hosted")
	githubClnt *util.GithubClient
)

const (
	istioDepsFile = "istio.deps"
	prTitlePrefix = "[DO NOT MERGE] Auto PR to update dependencies of "
	prBody        = "This PR will be merged automatically once checks are successful."
)

// Updates dependency objects in :deps to the latest stable version.
// Generates an MD5 digest of the latest dependencies, useful in avoiding making duplicate
// branches of the same code change.
// Returns a list of dependencies that were stale and have just been updated
func updateDepSHAGetFingerPrint(repo string, deps *[]dependency) (string, []dependency, error) {
	var depChangeList []dependency
	digest, err := githubClnt.GetHeadCommitSHA(repo, *baseBranch)
	if err != nil {
		return "", depChangeList, err
	}
	digest += *baseBranch + *hub
	for i, dep := range *deps {
		commitSHA, err := githubClnt.GetHeadCommitSHA(dep.RepoName, dep.ProdBranch)
		if err != nil {
			return "", depChangeList, err
		}
		digest += commitSHA
		if dep.LastStableSHA != commitSHA {
			(*deps)[i].LastStableSHA = commitSHA
			depChangeList = append(depChangeList, (*deps)[i])
		}

	}
	return util.GetMD5Hash(digest), depChangeList, nil
}

// Updates in the file all occurrences of the dependency identified by depName to
// a new reference ref. A reference could be a commit SHA, branch, or url.
func updateDepFile(file, depName, ref string) error {
	replaceReference := func(line *string, ref string) {
		idx := strings.Index(*line, "\"")
		*line = (*line)[:idx] + "\"" + ref + "\""
	}

	input, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	lines := strings.Split(string(input), "\n")
	found := false
	for i, line := range lines {
		if strings.Contains(line, depName+" = ") || strings.Contains(line, depName+"=") {
			replaceReference(&lines[i], ref)
			found = true
		}
	}
	if !found {
		return fmt.Errorf("no occurrence of %s found in %s", depName, file)
	}
	output := strings.Join(lines, "\n")
	return ioutil.WriteFile(file, []byte(output), 0600)
}

// Updates the list of dependencies in repo to the latest stable references
func updateDeps(repo string, deps *[]dependency, depChangeList *[]dependency) error {
	if repo != "istio" {
		for _, dep := range *deps {
			if err := updateDepFile(dep.File, dep.Name, dep.LastStableSHA); err != nil {
				return err
			}
		}
		return nil
	}
	args := ""
	for _, updatedDep := range *depChangeList {
		switch updatedDep.RepoName {
		case "mixer":
			args += fmt.Sprintf("-x %s,%s ", *hub, updatedDep.LastStableSHA)
		case "pilot":
			istioctlURL := fmt.Sprintf(
				"https://storage.googleapis.com/istio-artifacts/pilot/%s/artifacts/istioctl",
				updatedDep.LastStableSHA)
			args += fmt.Sprintf("-p %s,%s -i %s ", *hub, updatedDep.LastStableSHA, istioctlURL)
		case "auth":
			args += fmt.Sprintf("-c %s,%s ", *hub, updatedDep.LastStableSHA)
		default:
			return fmt.Errorf("unknown dependency: %s", updatedDep.Name)
		}
	}
	cmd := fmt.Sprintf("./install/updateVersion.sh %s", args)
	_, err := util.Shell(cmd)
	return err
}

// Deletes the local git repo just cloned
func cleanUp(repo string) error {
	if err := os.Chdir(".."); err != nil {
		return err
	}
	return os.RemoveAll(repo)
}

// Updates the given repository so that it uses the latest dependency references
// pushes new branch to remote, create pull request on master,
// which is auto-merged after presumbit
func updateDependenciesOf(repo string) error {
	log.Printf("Updating dependencies of %s\n", repo)
	if err := os.RemoveAll(repo); err != nil {
		return err
	}
	if _, err := util.ShellSilent("git clone " + githubClnt.Remote(repo)); err != nil {
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
	if _, err := util.Shell("git checkout " + *baseBranch); err != nil {
		return err
	}
	deps, err := deserializeDeps(istioDepsFile)
	if err != nil {
		return err
	}
	fingerPrint, depChangeList, err := updateDepSHAGetFingerPrint(repo, &deps)
	if err != nil {
		return err
	}
	if len(depChangeList) == 0 {
		log.Printf("%s is up to date. No commits are made.", repo)
		return nil
	}
	branch := "autoUpdateDeps_" + fingerPrint
	exists, err := githubClnt.ExistBranch(repo, branch)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("Branch already exists")
	}
	// if branch exists, stop here and do not create another PR of identical delta
	if err = githubClnt.CloseFailedPullRequests(
		prTitlePrefix, repo, *baseBranch); exists || err != nil {
		return err
	}
	if _, err := util.Shell("git checkout -b " + branch); err != nil {
		return err
	}
	if err := updateDeps(repo, &deps, &depChangeList); err != nil {
		return err
	}
	if err := serializeDeps(istioDepsFile, &deps); err != nil {
		return err
	}
	if _, err := util.Shell("git commit -am Update_Dependencies"); err != nil {
		return err
	}
	if _, err := util.Shell("git push --set-upstream origin " + branch); err != nil {
		return err
	}
	prTitle := prTitlePrefix + repo
	return githubClnt.CreatePullRequest(prTitle, prBody, branch, *baseBranch, repo)
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
	githubClnt, err = util.NewGithubClient(*owner, token)
	if err != nil {
		log.Panicf("Error when initializing github client: %v\n", err)
	}
	if *repo != "" { // only update dependencies of this repo
		if (*repo == "istio") == (*hub == "") {
			log.Printf("The hub flag hub must be set for istio/istio and must not for other repos\n")
			return
		}
		if err := updateDependenciesOf(*repo); err != nil {
			log.Printf("Failed to udpate dependency: %v\n", err)
		}
	} else { // update dependencies of all repos in the istio project
		repos, err := githubClnt.ListRepos()
		if err != nil {
			log.Printf("Error when fetching list of repos: %v\n", err)
			return
		}
		for _, r := range repos {
			if err := updateDependenciesOf(r); err != nil {
				log.Printf("Failed to udpate dependency: %v\n", err)
			}
		}
	}
}

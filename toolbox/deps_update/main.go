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
	"os/exec"
	"strings"

	"istio.io/test-infra/toolbox/util"
)

var (
	repo       = flag.String("repo", "", "Optional. Update dependencies of only this repository")
	owner      = flag.String("owner", "istio", "Github Owner or org")
	tokenFile  = flag.String("token_file", "", "File containing Github API Access Token")
	baseBranch = flag.String("base_branch", "master", "Branch from which the deps update commit is based")
	caHub      = flag.String("ca_hub", "", "Optional. Where new CA image is hosted")
	mixerHub   = flag.String("mixer_hub", "", "Optional. Where new mixer image is hosted")
	pilotHub   = flag.String("pilot_hub", "", "Optional. Where new pilot image is hosted")
	githubClnt *githubClient
)

const (
	istioDepsFile = "istio.deps"
	istioVersion  = "istio.VERSION"
)

// Generates an MD5 digest of the version set of the repo dependencies
// useful in avoiding making duplicate branches of the same code change
// Also updates dependency objects deserialized from istio.deps
func fingerPrintAndUpdateDepSHA(repo string, deps *[]dependency) (string, error) {
	digest, err := githubClnt.getHeadCommitSHA(repo, *baseBranch)
	if err != nil {
		return "", err
	}
	digest += *baseBranch + *caHub + *mixerHub + *pilotHub
	for i, dep := range *deps {
		commitSHA, err := githubClnt.getHeadCommitSHA(dep.RepoName, dep.ProdBranch)
		if err != nil {
			return "", err
		}
		digest += commitSHA
		(*deps)[i].LastStableSHA = commitSHA
	}
	return util.GetMD5Hash(digest), nil
}

// Reads the exported environment variable named :key in istio.VERSION
// and returns its value
func istioVersionGet(key string) (string, error) {
	// exec always forks a child process to execute the given command, which means
	// istio.VERSION is sourced to a forked bash process and all exported values
	// are gone once the child process exits.
	// The following is a workaround to pass value back to the parent process,
	// which is whoever executing this go script
	cmd := fmt.Sprintf("source istio.VERSION ; echo $%s", key)
	value, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	return strings.TrimSpace(string(value)), err
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

// Assumes at the root of istio directory
// Runs the updateVersion.sh script
func updateIstioDeps(oldPilot, newPilot string) error {
	update := func(depName string, newRef *string) error {
		if *newRef != "" {
			return updateDepFile(istioVersion, depName, *newRef)
		}
		return nil
	}

	if err := update("CA_HUB", caHub); err != nil {
		return err
	}
	if err := update("PILOT_HUB", pilotHub); err != nil {
		return err
	}
	if err := update("MIXER_HUB", mixerHub); err != nil {
		return err
	}
	// ISTIOCTL_URL has an embedded reference to pilot that will not be
	// updated by ./install/updateVersion.sh so must handle explicitly
	istioctlURL, err := istioVersionGet("ISTIOCTL_URL")
	if err != nil {
		return err
	}
	istioctlURL = strings.Replace(istioctlURL, oldPilot, newPilot, 1)
	cmd := fmt.Sprintf("./install/updateVersion.sh -i %s", istioctlURL)
	_, err = util.Shell(cmd)
	return err
}

// Updates the list of dependencies in repo to the latest stable references
func updateDeps(repo string, deps *[]dependency) error {
	updateDepFiles := func() error {
		for _, dep := range *deps {
			if err := updateDepFile(dep.File, dep.Name, dep.LastStableSHA); err != nil {
				return err
			}
		}
		return nil
	}
	getSHAByName := func(name string) (string, error) {
		for _, d := range *deps {
			if d.RepoName == name {
				return d.LastStableSHA, nil
			}
		}
		return "", fmt.Errorf("unknown dependency: %s", name)
	}

	if repo != "istio" {
		return updateDepFiles()
	}
	// read oldPilot before updateDepFiles changes it to newPilot
	oldPilot, err := istioVersionGet("PILOT_TAG")
	if err != nil {
		return err
	}
	if err = updateDepFiles(); err != nil {
		return err
	}
	newPilot, err := getSHAByName("pilot")
	if err != nil {
		return err
	}
	if err := updateIstioDeps(oldPilot, newPilot); err != nil {
		return err
	}
	return nil
	// TODO (chx) check outdated comments
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
	if _, err := util.ShellSilent("git clone " + githubClnt.remote(repo)); err != nil {
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
	depVersions, err := fingerPrintAndUpdateDepSHA(repo, &deps)
	if err != nil {
		return err
	}
	branch := "autoUpdateDeps_" + depVersions
	exists, err := githubClnt.existBranch(repo, branch)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("Branch already exists")
	}
	// if branch exists, stop here and do not create another PR of identical delta
	if err = githubClnt.closeFailedPullRequests(repo, *baseBranch); exists || err != nil {
		return err
	}
	if _, err := util.Shell("git checkout -b " + branch); err != nil {
		return err
	}
	if err := updateDeps(repo, &deps); err != nil {
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
	if err := githubClnt.createPullRequest(branch, *baseBranch, repo); err != nil {
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
		if *repo != "istio" && (*caHub != "" || *pilotHub != "" || *mixerHub != "") {
			log.Printf("The optional hub flags only apply to istio/istio\n")
			return
		}
		if err := updateDependenciesOf(*repo); err != nil {
			log.Printf("Failed to udpate dependency: %v\n", err)
		}
	} else { // update dependencies of all repos in the istio project
		repos, err := githubClnt.listRepos()
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

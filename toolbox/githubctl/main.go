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
	"log"
	"strings"

	u "istio.io/test-infra/toolbox/util"
)

var (
	owner       = flag.String("owner", "istio", "Github owner or org")
	tokenFile   = flag.String("token_file", "", "File containing Github API Access Token")
	op          = flag.String("op", "", "Operation to be performed")
	repo        = flag.String("repo", "", "Repository to which op is applied")
	baseBranch  = flag.String("base_branch", "", "Branch to which op is applied")
	refSHA      = flag.String("ref_sha", "", "Reference commit SHA used to update base branch")
	nextRelease = flag.String("next_release", "", "Tag of the next release")
	githubClnt  *u.GithubClient
)

const (
	istioDepsFile        = "istio.deps"
	releaseTagFile       = "istio.RELEASE"
	downloadScript       = "downloadIstio.sh"
	istioRepo            = "istio"
	masterBranch         = "master"
	dockerHub            = "docker.io/istio"
	logPrefix            = "Archives are available in "
	releasePRTtilePrefix = "[Auto Release] "
	releasePRBody        = "Update istio.VERSION and downloadIstio.sh"
)

// Panic if value not specified
func assertNotEmpty(name string, value *string) {
	if value == nil || *value == "" {
		log.Panicf("%s must be specified\n", name)
	}
}

func fastForward(repo, baseBranch, refSHA *string) error {
	assertNotEmpty("repo", repo)
	assertNotEmpty("base_branch", baseBranch)
	assertNotEmpty("ref_sha", refSHA)
	return githubClnt.FastForward(*repo, *baseBranch, *refSHA)
}

func createIstioReleaseTag() error {
	assertNotEmpty("base_branch", baseBranch)
	pickledDeps, err := githubClnt.GetFileContent(istioRepo, *baseBranch, istioDepsFile)
	if err != nil {
		return err
	}
	deps, err := u.DeserializeDepsFromString(pickledDeps)
	if err != nil {
		return err
	}
	releaseTag, err := githubClnt.GetFileContent(istioRepo, *baseBranch, releaseTagFile)
	if err != nil {
		return err
	}
	releaseTag = strings.Trim(releaseTag, "\n")
	releaseMsg := "Istio Release " + releaseTag
	for _, dep := range deps {
		if err := githubClnt.CreateAnnotatedTag(
			dep.RepoName, releaseTag, dep.LastStableSHA, releaseMsg); err != nil {
			return err
		}
	}
	return nil
}

func findPathToArchives(consoleLog *string) (string, error) {
	lines := strings.Split(*consoleLog, "\n")
	lastLine := lines[len(lines)-2]
	if strings.HasPrefix(lastLine, logPrefix) {
		dir := lastLine[len(logPrefix):]
		return dir, nil
	}
	return "", fmt.Errorf("failed to find path to release archives from console log")
}

func updateIstioVersion(releaseTag string) error {
	hubCommaTag := fmt.Sprintf("%s,%s", dockerHub, releaseTag)
	istioctl := fmt.Sprintf(
		"https://storage.googleapis.com/istio-artifacts/pilot/%s/artifacts/istioctl", releaseTag)
	cmd := fmt.Sprintf("./install/updateVersion.sh -p %s -c %s -x %s -i %s",
		hubCommaTag, hubCommaTag, hubCommaTag, istioctl)
	_, err := u.Shell(cmd)
	return err
}

func createIstioReleaseUploadArtifacts() error {
	assertNotEmpty("base_branch", baseBranch)
	assertNotEmpty("next_release", nextRelease)
	if err := u.CloneRepoCheckoutBranch(
		githubClnt, istioRepo, *baseBranch); err != nil {
		return err
	}
	defer func() {
		if err := u.RemoveLocalRepo(istioRepo); err != nil {
			log.Fatalf("Error during clean up: %v\n", err)
		}
	}()
	output, err := u.Shell("bash ./release/create_release_archives.sh")
	if err != nil {
		return err
	}
	archiveDir, err := findPathToArchives(&output)
	if err != nil {
		return err
	}
	releaseTag, err := u.ReadFile(releaseTagFile)
	if err != nil {
		return err
	}
	releaseTag = strings.Trim(releaseTag, "\n")
	if err := githubClnt.CreateReleaseUploadArchives(
		istioRepo, releaseTag, archiveDir); err != nil {
	}
	releaseBranch := "finalizeRelease-" + releaseTag
	if err := u.CheckoutNewBranchFromBaseBranch(
		masterBranch, releaseBranch); err != nil {
		return err
	}
	if err := updateIstioVersion(releaseTag); err != nil {
		return err
	}
	if err := u.WriteFile(releaseTagFile, *nextRelease); err != nil {
		return err
	}
	if err := u.UpdateKeyValueInFile(
		downloadScript, "ISTIO_VERSION", releaseTag); err != nil {
		return err
	}
	if err := u.CreateCommitPushToRemote(
		releaseBranch, releaseBranch); err != nil {
		return err
	}
	prTitle := releasePRTtilePrefix + releaseTag
	return githubClnt.CreatePullRequest(
		prTitle, releasePRBody, releaseBranch, *baseBranch, istioRepo)
}

func init() {
	flag.Parse()
	assertNotEmpty("token_file", tokenFile)
	token, err := u.GetAPITokenFromFile(*tokenFile)
	if err != nil {
		log.Panicf("Error accessing user supplied token_file: %v\n", err)
	}
	githubClnt = u.NewGithubClient(*owner, token)
}

func main() {
	switch *op {
	case "fastForward":
		if err := fastForward(repo, baseBranch, refSHA); err != nil {
			log.Printf("Error during fastForward: %v\n", err)
		}
	case "createIstioReleaseTag":
		if err := createIstioReleaseTag(); err != nil {
			log.Printf("Error during createIstioReleaseTag: %v\n", err)
		}
	case "tar":
		if err := createIstioReleaseUploadArtifacts(); err != nil {
			log.Printf("Error during createIstioReleaseUploadArtifacts: %v\n", err)
		}
	default:
		log.Printf("Unsupported operation: %s\n", *op)
	}
}

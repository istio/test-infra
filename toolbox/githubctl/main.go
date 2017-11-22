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
	"os"
	"strings"
	"time"

	u "istio.io/test-infra/toolbox/util"
)

var (
	owner                              = flag.String("owner", "istio", "Github owner or org")
	tokenFile                          = flag.String("token_file", "", "File containing Github API Access Token")
	op                                 = flag.String("op", "", "Operation to be performed")
	repo                               = flag.String("repo", "", "Repository to which op is applied")
	baseBranch                         = flag.String("base_branch", "", "Branch to which op is applied")
	refSHA                             = flag.String("ref_sha", "", "Reference commit SHA used to update base branch")
	nextRelease                        = flag.String("next_release", "", "Tag of the next release")
	hub                                = flag.String("hub", "", "Hub of the docker images")
	tag                                = flag.String("tag", "", "Tag of the release candidate")
	releaseOrg                         = flag.String("rel_org", "istio-releases", "GitHub Release Org")
	extraBranchesUpdateDownloadVersion = flag.String("update_rel_branches", "",
		"Extra branches where you want to update downloadIstioCandidate.sh, separated by comma")
	githubClnt *u.GithubClient
	ghClntRel  *u.GithubClient
)

const (
	// 0.2 release tooling
	istioVersionFile     = "istio.VERSION"
	istioDepsFile        = "istio.deps"
	releaseTagFile       = "istio.RELEASE"
	downloadScript       = "release/downloadIstioCandidate.sh"
	istioRepo            = "istio"
	masterBranch         = "master"
	export               = "export "
	dockerHub            = "docker.io/istio"
	releaseBaseDir       = "/tmp/release"
	releasePRTtilePrefix = "[Auto Release] "
	releaseBucketFmtStr  = "https://storage.googleapis.com/istio-release/releases/%s/%s"
	istioctlSuffix       = "istioctl"
	debianSuffix         = "deb"
	// release qualification trigger
	relQualificationPRTtilePrefix = "Release Qualification"
	greenBuildVersionFile         = "greenBuild.VERSION"
	dailyRepo                     = "daily-release"
)

func fastForward(repo, baseBranch, refSHA *string) error {
	u.AssertNotEmpty("repo", repo)
	u.AssertNotEmpty("base_branch", baseBranch)
	u.AssertNotEmpty("ref_sha", refSHA)
	isAncestor, err := githubClnt.SHAIsAncestorOfBranch(*repo, masterBranch, *refSHA)
	if err != nil {
		return err
	}
	if !isAncestor {
		log.Printf("SHA %s is not an ancestor of branch %s, resorts to no-op\n", *refSHA, masterBranch)
		return nil
	}
	return githubClnt.FastForward(*repo, *baseBranch, *refSHA)
}

func processIstioVersion(content *string) map[string]string {
	kv := make(map[string]string)
	lines := strings.Split(*content, "\n")
	for _, line := range lines {
		if idx := strings.Index(line, "="); idx != -1 {
			key := line[len(export):idx]
			value := strings.Trim(line[idx+1:], "\"")
			kv[key] = value
		}
	}
	return kv
}

func getReleaseTag() (string, error) {
	u.AssertNotEmpty("base_branch", baseBranch)
	releaseTag, err := githubClnt.GetFileContent(istioRepo, *baseBranch, releaseTagFile)
	if err != nil {
		return "", err
	}
	return strings.Trim(releaseTag, "\n"), nil
}

// TagIstioDepsForRelease creates release tag on each dependent repo of istio
func TagIstioDepsForRelease() error {
	u.AssertNotEmpty("base_branch", baseBranch)
	log.Printf("Fetching and processing istio.VERSION\n")
	istioVersion, err := githubClnt.GetFileContent(istioRepo, *baseBranch, istioVersionFile)
	if err != nil {
		return err
	}
	kv := processIstioVersion(&istioVersion)
	pickledDeps, err := githubClnt.GetFileContent(istioRepo, *baseBranch, istioDepsFile)
	if err != nil {
		return err
	}
	deps, err := u.DeserializeDepsFromString(pickledDeps)
	if err != nil {
		return err
	}
	log.Printf("Fetching release tag\n")
	releaseTag, err := getReleaseTag()
	if err != nil {
		return err
	}
	releaseMsg := "Istio Release " + releaseTag
	for _, dep := range deps {
		// use sha directly read from istio.VERSION in case updateVersion.sh was run
		// manually without also updating istio.deps
		log.Printf("Creating annotated tag [%s] on %s\n", releaseTag, dep.RepoName)
		ref, exists := kv[dep.Name]
		if !exists {
			return fmt.Errorf("ill-defined %s: unable to find %s", istioVersionFile, dep.Name)
		}
		// make sure ref is a SHA, special case where previous release is used in this release
		if u.ReleaseTagRegex.MatchString(ref) {
			ref, err = githubClnt.GetTagCommitSHA(dep.RepoName, ref)
			if err != nil {
				return err
			}
		}
		if err := githubClnt.CreateAnnotatedTag(
			dep.RepoName, releaseTag, ref, releaseMsg); err != nil {
			if strings.Contains(err.Error(), "Reference already exists") {
				log.Printf("Tag [%s] already exists on %s\n", releaseTag, dep.RepoName)
				prevTagSHA, err := githubClnt.GetTagCommitSHA(dep.RepoName, releaseTag)
				if err != nil {
					return err
				}
				if prevTagSHA == ref {
					log.Printf("Intended to tag [%s] at the same SHA, resort to no-op and continue\n", ref)
					continue
				} else {
					return fmt.Errorf("trying to tag [%s] at different SHA", ref)
				}
			}
		}
	}
	return nil
}

// UpdateIstioVersionAfterReleaseTagsMadeOnDeps runs updateVersion.sh to update
// istio.VERSION and then create a PR on istio
func UpdateIstioVersionAfterReleaseTagsMadeOnDeps() error {
	u.AssertNotEmpty("base_branch", baseBranch)
	releaseTag, err := getReleaseTag()
	if err != nil {
		return err
	}
	edit := func() error {
		hubCommaTag := fmt.Sprintf("%s,%s", dockerHub, releaseTag)
		istioctlURL := fmt.Sprintf(releaseBucketFmtStr, releaseTag, istioctlSuffix)
		debianURL := fmt.Sprintf(releaseBucketFmtStr, releaseTag, debianSuffix)
		cmd := fmt.Sprintf("./install/updateVersion.sh")
		// Auth
		cmd += fmt.Sprintf(" -c %s -A %s", hubCommaTag, debianURL)
		// Mixer
		cmd += fmt.Sprintf(" -x %s", hubCommaTag)
		// Pilot
		cmd += fmt.Sprintf(" -p %s -i %s -P %s", hubCommaTag, istioctlURL, debianURL)
		// Proxy
		cmd += fmt.Sprintf(" -r %s -E %s", releaseTag, debianURL)
		_, err := u.Shell(cmd)
		return err
	}
	releaseBranch := "Istio_Release_" + releaseTag
	body := "Update istio.Version"
	prTitle := releasePRTtilePrefix + body
	_, err = githubClnt.CreatePRUpdateRepo(releaseBranch, *baseBranch, istioRepo, prTitle, body, edit)
	return err
}

// CreateIstioReleaseUploadArtifacts creates a release on istio from the refSHA provided and uploads dependent artifacts
func CreateIstioReleaseUploadArtifacts() error {
	u.AssertNotEmpty("ref_sha", refSHA)
	u.AssertNotEmpty("base_branch", baseBranch)
	u.AssertNotEmpty("next_release", nextRelease)
	releaseTag, err := getReleaseTag()
	if releaseTag == *nextRelease {
		return fmt.Errorf("next_release should be greater than the current release")
	}
	if err != nil {
		return err
	}
	releaseBranch := "finalizeRelease-" + releaseTag
	prBody := fmt.Sprintf("Finalize release %s on istio", releaseTag)
	updateVersion := func() error {
		log.Printf("Updating downloadIstioCandidate.sh with latest release")
		return u.UpdateKeyValueInFile(
			downloadScript, "ISTIO_VERSION", fmt.Sprintf("${ISTIO_VERSION:-%s}", releaseTag))
	}
	edit := func() error {
		if _, err := u.Shell(
			"./release/create_release_archives.sh -d " + releaseBaseDir); err != nil {
			return err
		}
		archiveDir := releaseBaseDir + "/archives"
		if err := githubClnt.CreateReleaseUploadArchives(
			istioRepo, releaseTag, *refSHA, archiveDir); err != nil {
			return err
		}
		if err := u.WriteTextFile(releaseTagFile, *nextRelease); err != nil {
			return err
		}
		return updateVersion()
	}
	prTitle := releasePRTtilePrefix + prBody
	_, err = githubClnt.CreatePRUpdateRepo(releaseBranch, *baseBranch, istioRepo, prTitle, prBody, edit)
	if err != nil {
		return err
	}
	if *extraBranchesUpdateDownloadVersion != "" {
		extraBranches := strings.Split(*extraBranchesUpdateDownloadVersion, ",")
		for _, branch := range extraBranches {
			localBranch := fmt.Sprintf("%s-local", branch)
			if _, err := githubClnt.CreatePRUpdateRepo(localBranch, branch, istioRepo, prTitle, prBody, updateVersion); err != nil {
				// Only log out errors if failing update extra branches
				log.Printf("Warning! Failed to update downloadIstioCandidate.sh in branch %s", branch)
			}
		}
	}
	return nil
}

// DailyReleaseQualification triggers test jobs buy creating a PR that generates
// a GitHub notification. It blocks until PR status is known and returns nonzero
// value if failure. Links to test logs will also be logged to console.
func DailyReleaseQualification() error {
	u.AssertNotEmpty("hub", hub) // TODO (chx) default value of hub
	u.AssertNotEmpty("tag", tag)
	log.Printf("Creating PR to trigger release qualifications\n")
	prTitle := "[DO NOT MANUAL MERGE] " + relQualificationPRTtilePrefix + *refSHA
	prBody := fmt.Sprintf("Trigger release qualification jobs")
	timestamp := fmt.Sprintf("%v", time.Now().UnixNano())
	newBranch := "relQual_" + timestamp
	edit := func() error {
		if err := u.UpdateKeyValueInFile(greenBuildVersionFile, "HUB", *hub); err != nil {
			return err
		}
		if err := u.UpdateKeyValueInFile(greenBuildVersionFile, "TAG", *tag); err != nil {
			return err
		}
		if err := u.UpdateKeyValueInFile(greenBuildVersionFile, "TIME", timestamp); err != nil {
			return err
		}
		return nil
	}
	pr, err := ghClntRel.CreatePRUpdateRepo(newBranch, masterBranch, dailyRepo, prTitle, prBody, edit)
	if err != nil {
		return err
	}
	defer func() {
		log.Printf("Close the PR and delete its branch\n")
		if e := ghClntRel.ClosePRDeleteBranch(dailyRepo, pr); e != nil {
			log.Printf("Error in ClosePRDeleteBranch: %v\n", e)
		}
	}()

	verbose := true
	ci := u.NewCIState()
	retryDelay := 1 * time.Minute
	totalRetries := 60
	log.Printf("Waiting for all jobs starting. Results Polling starts in %v.\n", retryDelay)
	time.Sleep(retryDelay)
	err = u.Poll(retryDelay, totalRetries, func() (bool, error) {
		status, errPoll := ghClntRel.GetPRTestResults(dailyRepo, pr, verbose)
		verbose = false
		if errPoll != nil {
			return false, errPoll
		}
		exitPolling := false
		switch status {
		case ci.Success:
			log.Printf("Release qualification passed\n")
			exitPolling = true
		case ci.Failure:
			// All failures have been logged by GetPRTestResults()
			exitPolling = true
			errPoll = fmt.Errorf("release qualification failed")
		case ci.Pending:
			log.Printf("Results still pending. Will check again in %v.\n", retryDelay)
		}
		return exitPolling, errPoll
	})
	if err != nil { // qualification failed
		return err
	}
	log.Printf("Auto merging this PR to update daily release\n")
	return ghClntRel.MergePR(dailyRepo, pr)
}

func init() {
	flag.Parse()
	u.AssertNotEmpty("token_file", tokenFile)
	token, err := u.GetAPITokenFromFile(*tokenFile)
	if err != nil {
		log.Fatalf("Error accessing user supplied token_file: %v\n", err)
	}
	githubClnt = u.NewGithubClient(*owner, token)
	// a new github client is created for istio-releases org
	ghClntRel = u.NewGithubClient(*releaseOrg, token)
}

func main() {
	switch *op {
	case "fastForward":
		if err := fastForward(repo, baseBranch, refSHA); err != nil {
			log.Printf("Error during fastForward: %v\n", err)
		}
	case "tagIstioDepsForRelease":
		if err := TagIstioDepsForRelease(); err != nil {
			log.Printf("Error during TagIstioDepsForRelease: %v\n", err)
		}
	case "updateIstioVersion":
		if err := UpdateIstioVersionAfterReleaseTagsMadeOnDeps(); err != nil {
			log.Printf("Error during UpdateIstioVersionAfterReleaseTagsMadeOnDeps: %v\n", err)
		}
	case "uploadArtifacts":
		if err := CreateIstioReleaseUploadArtifacts(); err != nil {
			log.Printf("Error during CreateIstioReleaseUploadArtifacts: %v\n", err)
		}
	case "dailyRelQual":
		if err := DailyReleaseQualification(); err != nil {
			log.Printf("Error during DailyReleaseQualification: %v\n", err)
			os.Exit(1)
		}
	default:
		log.Printf("Unsupported operation: %s\n", *op)
	}
}

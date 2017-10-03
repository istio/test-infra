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
	releasePRBody        = "Update istio.VERSION and downloadIstioCandidate.sh"
	releaseBucketFmtStr  = "https://storage.googleapis.com/istio-release/releases/%s/%s"
	istioctlSuffix       = "istioctl"
	debianSuffix         = "deb"
)

// Exit if value not specified
func assertNotEmpty(name string, value *string) {
	if value == nil || *value == "" {
		log.Fatalf("%s must be specified\n", name)
	}
}

func fastForward(repo, baseBranch, refSHA *string) error {
	assertNotEmpty("repo", repo)
	assertNotEmpty("base_branch", baseBranch)
	assertNotEmpty("ref_sha", refSHA)
	isAncestor, err := githubClnt.SHAIsAncestorOfBranch(*repo, masterBranch, *refSHA)
	if err != nil {
		return err
	}
	if !isAncestor {
		log.Printf("SHA %s is not an ancestor of branch %s, resorts to no-op\n", refSHA, masterBranch)
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
	assertNotEmpty("base_branch", baseBranch)
	releaseTag, err := githubClnt.GetFileContent(istioRepo, *baseBranch, releaseTagFile)
	if err != nil {
		return "", err
	}
	return strings.Trim(releaseTag, "\n"), nil
}

// TagIstioDepsForRelease creates release tag on each dependent repo of istio
func TagIstioDepsForRelease() error {
	assertNotEmpty("base_branch", baseBranch)
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
					log.Printf("Intended to tag [%s] at the same SHA, resort to no-op and continue\n")
					continue
				} else {
					return fmt.Errorf("trying to tag [%s] at different SHA")
				}
			}
		}
	}
	return nil
}

func cloneIstioMakePR(newBranch, prTitle, prBody string, edit func() error) error {
	assertNotEmpty("base_branch", baseBranch)
	log.Printf("Cloning istio to local and checkout %s\n", *baseBranch)
	repoDir, err := u.CloneRepoCheckoutBranch(githubClnt, istioRepo, *baseBranch, newBranch)
	if err != nil {
		return err
	}
	defer func() {
		if err := u.RemoveLocalRepo(repoDir); err != nil {
			log.Fatalf("Error during clean up: %v\n", err)
		}
	}()
	if err := edit(); err != nil {
		return err
	}
	log.Printf("Staging commit and creating pull request\n")
	if err := u.CreateCommitPushToRemote(
		newBranch, newBranch); err != nil {
		return err
	}
	pr, err := githubClnt.CreatePullRequest(
		prTitle, prBody, newBranch, *baseBranch, istioRepo)
	if err != nil {
		return err
	}
	return githubClnt.AddlabelsToPR(istioRepo, pr, "release-note-none")
}

// UpdateIstioVersionAfterReleaseTagsMadeOnDeps runs updateVersion.sh to update
// istio.VERSION and then create a PR on istio
func UpdateIstioVersionAfterReleaseTagsMadeOnDeps() error {
	assertNotEmpty("base_branch", baseBranch)
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
		cmd += fmt.Sprintf(" -r %s -E %s", hubCommaTag, debianURL)
		_, err := u.Shell(cmd)
		return err
	}
	releaseBranch := "Istio_Release_" + releaseTag
	body := "Update istio.Version"
	prTitle := releasePRTtilePrefix + body
	return cloneIstioMakePR(releaseBranch, prTitle, body, edit)
}

// CreateIstioReleaseUploadArtifacts creates a release on istio and uploads dependent artifacts
func CreateIstioReleaseUploadArtifacts(refSHA string) error {
	assertNotEmpty("base_branch", baseBranch)
	assertNotEmpty("next_release", nextRelease)
	releaseTag, err := getReleaseTag()
	if releaseTag == *nextRelease {
		return fmt.Errorf("Next Release tag needs to be greater than the current release.")
	}
	if err != nil {
		return err
	}
	releaseBranch := "finalizeRelease-" + releaseTag
	prBody := fmt.Sprintf("Finalize release %s on istio", releaseTag)
	edit := func() error {
		if _, err := u.Shell(
			"./release/create_release_archives.sh -d " + releaseBaseDir); err != nil {
			return err
		}
		archiveDir := releaseBaseDir + "/archives"
		if err := githubClnt.CreateReleaseUploadArchives(
			istioRepo, releaseTag, refSHA, archiveDir); err != nil {
			return err
		}
		if err := u.WriteTextFile(releaseTagFile, *nextRelease); err != nil {
			return err
		}
		log.Printf("Updating download release candidate script with latest release")
		return u.UpdateKeyValueInFile(
			downloadScript, "ISTIO_VERSION", fmt.Sprintf("${ISTIO_VERSION:-%s}", releaseTag))
	}
	prTitle := releasePRTtilePrefix + prBody
	return cloneIstioMakePR(releaseBranch, prTitle, prBody, edit)
}

func init() {
	flag.Parse()
	assertNotEmpty("token_file", tokenFile)
	token, err := u.GetAPITokenFromFile(*tokenFile)
	if err != nil {
		log.Fatalf("Error accessing user supplied token_file: %v\n", err)
	}
	githubClnt = u.NewGithubClient(*owner, token)
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
		assertNotEmpty("ref_sha", refSHA)
		if err := CreateIstioReleaseUploadArtifacts(*refSHA); err != nil {
			log.Printf("Error during CreateIstioReleaseUploadArtifacts: %v\n", err)
		}
	default:
		log.Printf("Unsupported operation: %s\n", *op)
	}
}

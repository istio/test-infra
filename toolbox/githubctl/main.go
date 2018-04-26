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
	"time"

	s "istio.io/test-infra/sisyphus"
	u "istio.io/test-infra/toolbox/util"
)

var (
	owner      = flag.String("owner", "istio", "Github owner or org")
	tokenFile  = flag.String("token_file", "", "File containing Github API Access Token")
	op         = flag.String("op", "", "Operation to be performed")
	repo       = flag.String("repo", "", "Repository to which op is applied")
	baseBranch = flag.String("base_branch", "", "Branch to which op is applied")
	refSHA     = flag.String("ref_sha", "", "Reference commit SHA used to update base branch")
	hub        = flag.String("hub", "", "Hub of the docker images")
	tag        = flag.String("tag", "", "Tag of the release candidate")
	releaseOrg = flag.String("rel_org", "istio-releases", "GitHub Release Org")
	gcsPath    = flag.String("gcs_path", "", "The path to the GCS bucket")
	githubClnt *u.GithubClient
	ghClntRel  *u.GithubClient
	// unable to query post-submit jobs as GitHub is unaware of them
	// needs to be consistent with prow config map
	postSubmitJobs = []string{
		"istio-postsubmit",
		"e2e-suite-rbac-no_auth",
		"e2e-suite-rbac-auth",
		"e2e-cluster_wide-auth",
	}
)

const (
	masterBranch = "master"
	// Prow
	prowProject   = "istio-testing"
	prowZone      = "us-west1-a"
	gubernatorURL = "https://k8s-gubernator.appspot.com/build/istio-prow"
	gcsBucket     = "istio-prow"
	// latest green SHA
	maxCommitDepth = 10
	maxRunDepth    = 100
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

// TODO (chx) caching scheme to reduce IO to GCS. Also use go routines to increase parallelism
func isJobSuccessOnSHA(sha, job string, prowAccessor *s.ProwAccessor) (bool, error) {
	runNumber, err := prowAccessor.GetLatestRun(job)
	if err != nil {
		return false, fmt.Errorf("failed to get latest run number of %s: %v", job, err)
	}
	for i := 0; i < maxRunDepth; i++ {
		result, err := prowAccessor.GetResult(job, runNumber)
		if err != nil {
			return false, fmt.Errorf(
				"failed to get result of %s at run number %d: %v", job, runNumber, err)
		}
		if result.SHA == sha {
			return result.Passed, nil
		}
	}
	return false, fmt.Errorf(
		"Did not find matching sha [%s] for job %s after checking latest %d runs", sha, job, maxRunDepth)
}

func getLatestGreenSHA() (string, error) {
	u.AssertNotEmpty("repo", repo)
	u.AssertNotEmpty("base_branch", baseBranch)
	gcsClient := u.NewGCSClient(gcsBucket)
	prowAccessor := s.NewProwAccessor(prowProject, prowZone, gubernatorURL, gcsBucket, gcsClient)
	sha, err := githubClnt.GetHeadCommitSHA(*repo, *baseBranch)
	if err != nil {
		log.Fatalf("failed to get the head commit sha of %s/%s", *repo, *baseBranch)
	}
	for i := 0; i < maxCommitDepth; i++ {
		allChecksPassed := true
		for _, postSubmitJob := range postSubmitJobs {
			passed, err := isJobSuccessOnSHA(sha, postSubmitJob, prowAccessor)
			if err != nil {
				log.Fatalf("failed to check if sha %s passed job %s", sha, postSubmitJob)
			}
			if !passed {
				allChecksPassed = false
				break
			}
		}
		if allChecksPassed {
			return sha, nil
		}
		parentSHA, err := githubClnt.GetParentSHA(*repo, *baseBranch, sha)
		if err != nil {
			log.Fatalf("failed to find the parent sha of %s in %s/%s", sha, *repo, *baseBranch)
		}
		sha = parentSHA
	}
	return "", fmt.Errorf("exceeded max commit depth")
}

// DailyReleaseQualification triggers test jobs buy creating a PR that generates
// a GitHub notification. It blocks until PR status is known and returns nonzero
// value if failure. Links to test logs will also be logged to console.
func DailyReleaseQualification(baseBranch *string) error {
	u.AssertNotEmpty("hub", hub) // TODO (chx) default value of hub
	u.AssertNotEmpty("tag", tag)
	u.AssertNotEmpty("gcs_path", gcsPath)
	var dstBranch string
	// we could have made baseBranch have a default value, but that breaks all the places
	// where baseBranch must be passed in cmdline and a default value is not acceptable
	// therefore, if a branch is not passed in use masterBranch as the default destination
	if baseBranch != nil {
		dstBranch = *baseBranch
	} else {
		dstBranch = masterBranch
	}
	log.Printf("Creating PR to trigger release qualifications on %s branch\n", dstBranch)
	prTitle := fmt.Sprintf("%s - %s", relQualificationPRTtilePrefix, *tag)
	prBody := "This is a generated PR that triggers release qualification tests, and will be automatically merged " +
		"if all tests pass. In case some test fails, you can manually rerun the failing tests using /test. Force " +
		"merging this PR will suppress the test failures and let the release pipeline continue."
	timestamp := fmt.Sprintf("%v", time.Now().UnixNano())
	srcBranch := "relQual_" + timestamp
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
		if err := u.UpdateKeyValueInFile(greenBuildVersionFile, "ISTIO_REL_URL",
			fmt.Sprintf("https://storage.googleapis.com/%s", *gcsPath)); err != nil {
			return err
		}
		return nil
	}
	pr, err := ghClntRel.CreatePRUpdateRepo(srcBranch, dstBranch, dailyRepo, prTitle, prBody, edit)
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
	retryDelay := 5 * time.Minute
	maxWait := 20 * time.Hour
	totalRetries := int(maxWait / retryDelay)
	log.Printf("Waiting for all jobs starting. Results Polling starts in %v.\n", retryDelay)
	time.Sleep(retryDelay)

	err = u.Poll(retryDelay, totalRetries, func() (bool, error) {
		pr, err = ghClntRel.GetPR(dailyRepo, *pr.Number)
		if err != nil {
			return true, err
		}
		if *pr.Merged {
			// PR is apparently closed manually. Exit the loop.
			log.Printf("pr was manually merged.\n")
			return true, nil
		}
		if *pr.State == "closed" {
			// PR is apparently closed manually. Exit the loop.
			return false, fmt.Errorf("pr close was manually closed")
		}

		status, errPoll := ghClntRel.GetPRTestResults(dailyRepo, pr, verbose)
		verbose = false
		if errPoll != nil {
			return false, errPoll
		}
		exitPolling := false
		switch status {
		case ci.Success:
			exitPolling = true
			log.Printf("Auto merging this PR to update daily release\n")
			errPoll = ghClntRel.MergePR(dailyRepo, pr)
			*pr.Merged = true

		case ci.Pending:
			log.Printf("Results still pending. Will check again in %v.\n", retryDelay)
		case ci.Error:
		case ci.Failure:
			// Go back to sleep until timeout, so that release engineer can potentially suppress test failure or retest
			// in github directly.
		}
		return exitPolling, errPoll
	})
	// Fail to poll or merge
	if err != nil {
		return err
	}
	// In case the PR is merged manually or automatically
	if *pr.Merged {
		return nil
	}
	return fmt.Errorf("release qualification failed")
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
	case "dailyRelQual":
		if err := DailyReleaseQualification(baseBranch); err != nil {
			log.Printf("Error during DailyReleaseQualification: %v\n", err)
			os.Exit(1)
		}
	case "getLatestGreenSHA":
		latestGreenSHA, err := getLatestGreenSHA()
		if err != nil {
			log.Printf("Error during getLatestGreenSHA: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s", latestGreenSHA)
	default:
		log.Printf("Unsupported operation: %s\n", *op)
	}
}

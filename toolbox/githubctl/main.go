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
	"os"
	"sync"
	"time"

	"github.com/golang/glog"

	s "istio.io/test-infra/sisyphus"
	u "istio.io/test-infra/toolbox/util"
)

var (
	owner                 = flag.String("owner", "istio", "Github owner or org")
	tokenFile             = flag.String("token_file", "", "File containing Github API Access Token")
	op                    = flag.String("op", "", "Operation to be performed")
	repo                  = flag.String("repo", "", "Repository to which op is applied")
	baseBranch            = flag.String("base_branch", "", "Branch to which op is applied")
	refSHA                = flag.String("ref_sha", "", "Reference commit SHA used to update base branch")
	hub                   = flag.String("hub", "", "Hub of the docker images")
	tag                   = flag.String("tag", "", "Tag of the release candidate")
	releaseOrg            = flag.String("rel_org", "istio-releases", "GitHub Release Org")
	gcsPath               = flag.String("gcs_path", "", "The path to the GCS bucket")
	maxCommitDepth        = flag.Int("max_commit_depth", 200, "Max number of commits before HEAD to check if green")
	maxRunDepth           = flag.Int("max_run_depth", 500, "Max number of runs before the latest one of which results are checked")
	maxConcurrentRequests = flag.Int("max_concurrent_reqs", 50, "Max number of concurrent requests permitted")
	githubClnt            *u.GithubClient
	ghClntRel             *u.GithubClient
	// unable to query post-submit jobs as GitHub is unaware of them
	// needs to be consistent with prow config map
	postSubmitJobsMap = map[string][]string{
		"master": {
			"e2e-mixer-no_auth",
			"e2e-bookInfoTests-envoyv2-v1alpha3",
			"istio-pilot-e2e-envoyv2-v1alpha3",
			"e2e-simpleTests",
			"e2e-dashboard",
			"istio-postsubmit",
		},
		"release-1.0.0-snapshot-0": {
			"e2e-mixer-no_auth",
			"e2e-bookInfoTests-envoyv2-v1alpha3",
			"istio-pilot-e2e-envoyv2-v1alpha3",
			"e2e-simpleTests",
			"e2e-dashboard",
			"istio-postsubmit",
		},
		"release-0.8": {
			"istio-postsubmit",
			"e2e-suite-rbac-no_auth",
			"e2e-suite-rbac-auth",
			"e2e-cluster_wide-auth",
		},
	}
)

const (
	masterBranch = "master"
	// Prow
	prowProject   = "istio-testing"
	prowZone      = "us-west1-a"
	gubernatorURL = "https://k8s-gubernator.appspot.com/build/istio-prow"
	gcsBucket     = "istio-prow"
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
		glog.Infof("SHA %s is not an ancestor of branch %s, resorts to no-op\n", *refSHA, masterBranch)
		return nil
	}
	return githubClnt.FastForward(*repo, *baseBranch, *refSHA)
}

type task struct {
	job       string
	runNumber int
}

// preprocessProwResults downloads the most recent prow results up to maxRunDepth
// then returns a two-level map job -> sha -> passed (true) or failed (false)
func preprocessProwResults() map[string]map[string]bool {
	glog.Infof("Start preprocessing prow results")
	prowAccessor := s.NewProwAccessor(
		prowProject,
		prowZone,
		gubernatorURL,
		gcsBucket,
		u.NewGCSClient(gcsBucket))
	cache := make(map[string]map[string]bool)
	tasksCh := make(chan *task, *maxConcurrentRequests)
	var wg sync.WaitGroup
	mutex := &sync.Mutex{}
	for i := 0; i < *maxConcurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				t, more := <-tasksCh
				if !more {
					break
				}
				result, err := prowAccessor.GetResult(t.job, t.runNumber)
				if err != nil {
					glog.V(1).Infof("failed to get result of %s at run number %d. Skip.", t.job, t.runNumber)
					continue
				}
				mutex.Lock()
				cache[t.job][result.SHA] = result.Passed
				mutex.Unlock()
			}
		}()
	}
	postSubmitJobs := postSubmitJobsMap[*baseBranch]
	// note: if postSubmitJobs was not found in map, the for loop exits immediately
	for _, job := range postSubmitJobs {
		cache[job] = make(map[string]bool)
		runNumber, err := prowAccessor.GetLatestRun(job)
		if err != nil {
			glog.Fatalf("failed to get latest run number of %s: %v", job, err)
		}
		// download the most recent prow results up to maxRunDepth
		for i := 0; i < *maxRunDepth; i++ {
			tasksCh <- &task{job, runNumber}
			runNumber--
		}
	}
	close(tasksCh)
	wg.Wait()
	return cache
}

func getLatestGreenSHA() (string, error) {
	u.AssertNotEmpty("repo", repo)
	u.AssertNotEmpty("base_branch", baseBranch)
	u.AssertPositive("max_commit_depth", maxCommitDepth)
	u.AssertPositive("max_run_depth", maxRunDepth)
	results := preprocessProwResults()
	sha, err := githubClnt.GetHeadCommitSHA(*repo, *baseBranch)
	if err != nil {
		glog.Fatalf("failed to get the head commit sha of %s/%s: %v", *repo, *baseBranch, err)
	}
	postSubmitJobs, found := postSubmitJobsMap[*baseBranch]
	if !found {
		return "", fmt.Errorf("cannot find post submit jobs for branch %s", *baseBranch)
	}
	for i := 0; i < *maxCommitDepth; i++ {
		glog.Infof("Checking if [%s] passed all checks. %d commits before HEAD", sha, i)
		allChecksPassed := true
		for _, job := range postSubmitJobs {
			passed, keyExists := results[job][sha]
			if !keyExists {
				glog.V(1).Infof("Results unknown in local cache for [%s] at [%s], treat the test as failed", job, sha)
			}
			if !passed {
				glog.Infof("[%s] failed on [%s]", sha, job)
				allChecksPassed = false
			}
		}
		if allChecksPassed {
			glog.Infof("Found latest green sha [%s] for %s/%s", sha, *repo, *baseBranch)
			return sha, nil
		}
		parentSHA, err := githubClnt.GetParentSHA(*repo, *baseBranch, sha)
		if err != nil {
			glog.Fatalf("failed to find the parent sha of %s in %s/%s", sha, *repo, *baseBranch)
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
	if baseBranch != nil && len(*baseBranch) != 0 {
		dstBranch = *baseBranch
	} else {
		dstBranch = masterBranch
	}
	glog.Infof("Creating PR to trigger release qualifications on %s branch\n", dstBranch)
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
		glog.Infof("Close the PR and delete its branch\n")
		if e := ghClntRel.ClosePRDeleteBranch(dailyRepo, pr); e != nil {
			glog.Infof("Error in ClosePRDeleteBranch: %v\n", e)
		}
	}()

	retryDelay := 5 * time.Minute
	maxWait := 20 * time.Hour
	totalRetries := int(maxWait / retryDelay)
	glog.Infof("Waiting for all jobs starting. Results Polling starts in %v.\n", retryDelay)
	time.Sleep(retryDelay)

	err = u.Poll(retryDelay, totalRetries, func() (bool, error) {
		pr, err = ghClntRel.GetPR(dailyRepo, *pr.Number)
		if err != nil {
			return true, err
		}
		if *pr.Merged {
			// PR is already merged. Either all tests have passed or it is manually merged. Exit the loop.
			glog.Infof("pr was has been merged.\n")
			return true, nil
		}
		if *pr.State == "closed" {
			// PR is closed and not merged, which is a signal the qualification is abandoned.
			glog.Infof("pr has been closed.\n")
			return true, nil
		}
		return false, nil
	})
	// Fail to poll
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
		glog.Fatalf("Error accessing user supplied token_file: %v\n", err)
	}
	githubClnt = u.NewGithubClient(*owner, token)
	// a new github client is created for istio-releases org
	ghClntRel = u.NewGithubClient(*releaseOrg, token)
}

func main() {
	switch *op {
	case "fastForward":
		if err := fastForward(repo, baseBranch, refSHA); err != nil {
			glog.Infof("Error during fastForward: %v\n", err)
		}
	case "dailyRelQual":
		if err := DailyReleaseQualification(baseBranch); err != nil {
			glog.Infof("Error during DailyReleaseQualification: %v\n", err)
			os.Exit(1)
		}
	case "getLatestGreenSHA":
		latestGreenSHA, err := getLatestGreenSHA()
		if err != nil {
			glog.Infof("Error during getLatestGreenSHA: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s", latestGreenSHA)
	default:
		glog.Infof("Unsupported operation: %s\n", *op)
	}
}

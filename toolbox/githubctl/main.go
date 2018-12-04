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
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/test-infra/prow/config"

	s "istio.io/test-infra/sisyphus"
	u "istio.io/test-infra/toolbox/util"
)

var (
	owner                 = flag.String("owner", "istio", "Github owner or org")
	tokenFile             = flag.String("token_file", "", "File containing Github API Access Token.")
	op                    = flag.String("op", "", "Operation to be performed")
	repo                  = flag.String("repo", "", "Repository to which op is applied")
	pipelineType          = flag.String("pipeline", "", "Pipeline type daily/monthly")
	baseBranch            = flag.String("base_branch", "", "Branch to which op is applied")
	refSHA                = flag.String("ref_sha", "", "Commit SHA used by the operation")
	hub                   = flag.String("hub", "", "Hub of the docker images")
	tag                   = flag.String("tag", "", "Tag of the release candidate")
	releaseOrg            = flag.String("rel_org", "istio-releases", "GitHub Release Org")
	gcsPath               = flag.String("gcs_path", "", "The path to the GCS bucket")
	skip                  = flag.String("skip", "", "comma separated list of jobs to skip")
	prNum                 = flag.Int("pr_num", 0, "PR number")
	maxCommitDepth        = flag.Int("max_commit_depth", 200, "Max number of commits before HEAD to check if green")
	maxRunDepth           = flag.Int("max_run_depth", 500, "Max number of runs before the latest one of which results are checked")
	maxConcurrentRequests = flag.Int("max_concurrent_reqs", 50, "Max number of concurrent requests permitted")
	githubClnt            *u.GithubClient
	ghClntRel             *u.GithubClient
)

const (
	masterBranch = "master"
	// Prow
	prowProject   = "istio-testing"
	prowZone      = "us-west1-a"
	gubernatorURL = "https://k8s-gubernator.appspot.com/build/istio-prow"
	gcsBucket     = "istio-prow"
	// release pipeline triggers
	relBuildPRTtilePrefix         = " - Build"
	relQualificationPRTtilePrefix = " - Qualification"
	relReleasePRTtilePrefix       = " - Release"
	greenBuildVersionFile         = "test/greenBuild.VERSION"
	createBuildParametersCmd      = "./rel_scripts/create_release_build_parameters.sh -b %s -p %s -v %s"
	copyEnvToTestCmd              = "cp %s/build/build_parameters.sh %s/test/build_parameters.sh"
	copyEnvToReleaseCmd           = "cp %s/test/build_parameters.sh %s/release/build_parameters.sh"
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

func getBaseSha(repo *string, prNumber int) (string, error) {
	u.AssertNotEmpty("repo", repo)
	pr, err := githubClnt.GetPR(*repo, prNumber)
	if err != nil {
		return "", err
	}
	return *pr.Base.SHA, nil
}

type task struct {
	job       string
	runNumber int
}

// preprocessProwResults downloads the most recent prow results up to maxRunDepth
// then returns a two-level map job -> sha -> passed (true) or failed (false)
func preprocessProwResults(postSubmitJobs map[string]struct{}) map[string]map[string]bool {
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
	// if postSubmitJobs is empty, the for loop exits immediately
	for job := range postSubmitJobs {
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
	postSubmitJobs := readPostsubmitListFromProwConfig(*owner, *repo, *baseBranch)
	skipJobs := strings.Split(*skip, ",")
	for _, skipedJob := range skipJobs {
		delete(postSubmitJobs, skipedJob)
	}
	results := preprocessProwResults(postSubmitJobs)
	sha, err := githubClnt.GetHeadCommitSHA(*repo, *baseBranch)
	if err != nil {
		glog.Fatalf("failed to get the head commit sha of %s/%s: %v", *repo, *baseBranch, err)
	}
	for i := 0; i < *maxCommitDepth; i++ {
		glog.Infof("Checking if [%s] passed all checks. %d commits before HEAD", sha, i)
		allChecksPassed := true
		for job := range postSubmitJobs {
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

// ReleasePipelineBuild triggers build job by creating a PR that generates GitHub notification.
func ReleasePipelineBuild(baseBranch *string) error {
	u.AssertNotEmpty("pipeline", pipelineType)
	u.AssertNotEmpty("tag", tag)
	u.AssertNotEmpty("base_branch", baseBranch)
	dstBranch := *baseBranch
	glog.Infof("Creating PR to trigger build on %s branch\n", dstBranch)
	prTitle := *tag + relBuildPRTtilePrefix
	prBody := "This is a generated PR that triggers release build, and will be automatically merged "
	timestamp := fmt.Sprintf("%v", time.Now().UnixNano())
	srcBranch := "relQual_" + timestamp
	edit := func() error {
		createParametersCmd := fmt.Sprintf(createBuildParametersCmd, dstBranch, *pipelineType, *tag)
		glog.Infof("Running cmd: %s", createParametersCmd)
		_, err := u.Shell(createParametersCmd)
		return err
	}
	_, err := ghClntRel.CreatePRUpdateRepo(srcBranch, dstBranch, dailyRepo, prTitle, prBody, edit)
	return err
}

// ReleasePipelineQualification triggers test jobs buy creating a PR that generates
// a GitHub notification.
func ReleasePipelineQualification(baseBranch *string) error {
	u.AssertNotEmpty("tag", tag)
	u.AssertNotEmpty("pipeline", pipelineType)
	u.AssertNotEmpty("base_branch", baseBranch)
	dstBranch := *baseBranch
	glog.Infof("Creating PR to trigger release qualifications on %s branch\n", dstBranch)
	prTitle := *tag + relQualificationPRTtilePrefix
	prBody := "This is a generated PR that triggers release qualification tests, and will be automatically merged " +
		"if all tests pass. In case some test fails, you can manually rerun the failing tests using /test. Force " +
		"merging this PR will suppress the test failures and let the release pipeline continue."
	timestamp := fmt.Sprintf("%v", time.Now().UnixNano())
	srcBranch := "relQual_" + timestamp
	edit := func() error {
		copyCmd := fmt.Sprintf(copyEnvToTestCmd, *pipelineType, *pipelineType)
		_, err := u.Shell(copyCmd)
		return err
	}
	_, err := ghClntRel.CreatePRUpdateRepo(srcBranch, dstBranch, dailyRepo, prTitle, prBody, edit)
	return err
}

// ReleasePipelineRelease triggers release job for finishing release pipeline by creating a PR
//  that generates a GitHub notification.
func ReleasePipelineRelease(baseBranch *string) error {
	u.AssertNotEmpty("pipeline", pipelineType)
	u.AssertNotEmpty("tag", tag)
	u.AssertNotEmpty("base_branch", baseBranch)
	dstBranch := *baseBranch
	glog.Infof("Creating PR to trigger release on %s branch\n", dstBranch)
	prTitle := *tag + relReleasePRTtilePrefix
	prBody := "This is a generated PR that triggers release job, and will be automatically merged "
	timestamp := fmt.Sprintf("%v", time.Now().UnixNano())
	srcBranch := "relRelease_" + timestamp
	edit := func() error {
		copyCmd := fmt.Sprintf(copyEnvToReleaseCmd, *pipelineType, *pipelineType)
		_, err := u.Shell(copyCmd)
		return err
	}
	_, err := ghClntRel.CreatePRUpdateRepo(srcBranch, dstBranch, dailyRepo, prTitle, prBody, edit)
	return err
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
	prTitle := *tag + relQualificationPRTtilePrefix
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
		if err := u.UpdateKeyValueInFile(greenBuildVersionFile, "SHA", *refSHA); err != nil {
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

func traverseJobTree(job config.Postsubmit, postsubmitJobs map[string]struct{}, targetBranch string) {
	if job.Brancher.RunsAgainstBranch(targetBranch) {
		postsubmitJobs[job.Name] = struct{}{}
		if len(job.RunAfterSuccess) > 0 {
			for _, childJob := range job.RunAfterSuccess {
				traverseJobTree(childJob, postsubmitJobs, targetBranch)
			}
		}
	}
}

func readPostsubmitListFromProwConfig(org, repo, branch string) map[string]struct{} {
	postsubmitJobs := make(map[string]struct{})
	repoRootDir, err := u.Shell("git rev-parse --show-toplevel")
	repoRootDir = strings.TrimSuffix(repoRootDir, "\n")
	if err != nil {
		glog.Fatalf("cannot find repo root directory path")
	}
	prowConfigYaml := filepath.Join(repoRootDir, "prow/config.yaml")
	prowConfig, err := config.Load(prowConfigYaml, "")
	if err != nil {
		glog.Fatalf("could not read configs: %v", err)
	}

	for _, job := range prowConfig.Postsubmits[fmt.Sprintf("%s/%s", org, repo)] {
		traverseJobTree(job, postsubmitJobs, branch)
	}
	return postsubmitJobs
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
	// the following three cases are related to release pipeline
	case "relPipelineBuild":
		if err := ReleasePipelineBuild(baseBranch); err != nil {
			glog.Infof("Error during ReleasePipelineBuild: %v\n", err)
			os.Exit(1)
		}
	case "relPipelineQual":
		if err := ReleasePipelineQualification(baseBranch); err != nil {
			glog.Infof("Error during ReleasePipelineQualification: %v\n", err)
			os.Exit(1)
		}
	case "relPipelineRelease":
		if err := ReleasePipelineRelease(baseBranch); err != nil {
			glog.Infof("Error during ReleasePipelineRelease: %v\n", err)
			os.Exit(1)
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
	case "getBaseSHA":
		baseSha, err := getBaseSha(repo, *prNum)
		if err != nil {
			glog.Info(err)
			os.Exit(1)
		}
		fmt.Print(baseSha)
	default:
		glog.Infof("Unsupported operation: %s\n", *op)
	}
}

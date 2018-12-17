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
	"strings"
	"time"

	"github.com/golang/glog"

	u "istio.io/test-infra/toolbox/util"
)

var (
	owner        = flag.String("owner", "istio", "Github owner or org")
	tokenFile    = flag.String("token_file", "", "File containing Github API Access Token.")
	op           = flag.String("op", "", "Operation to be performed")
	repo         = flag.String("repo", "", "Repository to which op is applied")
	pipelineType = flag.String("pipeline", "", "Pipeline type daily/monthly")
	baseBranch   = flag.String("base_branch", "", "Branch to which op is applied")
	refSHA       = flag.String("ref_sha", "", "Commit SHA used by the operation")
	tag          = flag.String("tag", "", "Tag of the release candidate")
	prNum        = flag.Int("pr_num", 0, "PR number")

	githubClnt *u.GithubClient
)

const (
	masterBranch = "master"
	retest       = "/retest"
	maxRetests   = 3
)

func fastForward(repo, baseBranch, refSHA string) error {
	isAncestor, err := githubClnt.SHAIsAncestorOfBranch(repo, masterBranch, refSHA)
	if err != nil {
		return err
	}
	if !isAncestor {
		glog.Infof("SHA %s is not an ancestor of branch %s, resorts to no-op\n", refSHA, masterBranch)
		return nil
	}
	return githubClnt.FastForward(repo, baseBranch, refSHA)
}

func getBaseSha(repo string, prNumber int) (string, error) {
	pr, err := githubClnt.GetPR(repo, prNumber)
	if err != nil {
		return "", err
	}
	return *pr.Base.SHA, nil
}

// CreateReleaseRequest triggers release pipeline by creating a PR.
func CreateReleaseRequest(repo, pipelineType, tag, branch, sha string) error {
	glog.Infof("Creating PR to trigger build on %s branch\n", branch)
	prTitle := fmt.Sprintf("%s %s", strings.ToUpper(pipelineType), tag)
	prBody := "This is a generated PR that triggers a release, and will be automatically merged when all required tests have passed."
	timestamp := fmt.Sprintf("%v", time.Now().UnixNano())
	srcBranch := "release_" + timestamp
	edit := func() error {
		f, err := os.Create("./" + pipelineType + "/release_params.sh")
		if err != nil {
			return err
		}
		defer func() {
			cerr := f.Close()
			if err != nil {
				glog.Info(cerr)
			}
		}()

		if _, err := f.WriteString("export CB_BRANCH=" + branch + "\n"); err != nil {
			return err
		}
		if _, err := f.WriteString("export CB_PIPELINE_TYPE=" + pipelineType + "\n"); err != nil {
			return err
		}
		if _, err := f.WriteString("export CB_VERSION=" + tag + "\n"); err != nil {
			return err
		}
		if _, err := f.WriteString("export CB_COMMIT=" + sha + "\n"); err != nil {
			return err
		}
		return nil
	}
	_, err := githubClnt.CreatePRUpdateRepo(srcBranch, branch, repo, prTitle, prBody, edit)
	return err
}

// CleanupReleaseRequests merges tested release requests, and close the expired ones (not passing)
func CleanupReleaseRequests(owner, repo string) error {
	pullQueries := []string{
		fmt.Sprintf("repo:%s/%s", owner, repo),
		"type:pr",
		"is:open",
	}

	allPulls, err := githubClnt.SearchIssues(pullQueries, "created", "desc")
	if err != nil {
		return err
	}
	utc, _ := time.LoadLocation("UTC")
	for _, pull := range allPulls {
		pr, err := githubClnt.GetPR(repo, *pull.Number)
		if err != nil {
			return err
		}

		// Close the PR if it is expired (after 1 day)
		expiresAt := pr.CreatedAt.In(utc).Add(24 * time.Hour)
		if time.Now().In(utc).After(expiresAt) {
			if err2 := githubClnt.CreateComment(repo, pull, "Tests did not pass and this request has expired. Closing out."); err != nil {
				return err2
			}
			if err2 := githubClnt.ClosePRDeleteBranch(repo, pr); err != nil {
				return err2
			}
			glog.Infof("Closed https://github.com/%s/%s/pull/%d.", owner, repo, *pr.Number)
			break
		}

		status, err := githubClnt.GetPRTestResults(repo, pr, true)
		if err != nil {
			return err
		}
		ci := u.NewCIState()
		switch status {
		case ci.Success:
			if err := githubClnt.MergePR(repo, pr); err != nil {
				return err
			}
			if err := githubClnt.ClosePRDeleteBranch(repo, pr); err != nil {
				return err
			}
			glog.Infof("Merged https://github.com/%s/%s/pull/%d. Skipping.", owner, repo, *pr.Number)

		case ci.Pending:
			glog.Infof("https://github.com/%s/%s/pull/%d is still being tested. Skipping.", owner, repo, *pr.Number)
		case ci.Error:
		case ci.Failure:
			// Trigger a retest
			comments, err := githubClnt.ListIssueComments(repo, pull)
			if err != nil {
				return err
			}
			retestCount := 0
			for _, comment := range comments {
				if *comment.Body == retest {
					retestCount++
				}
			}
			if retestCount < maxRetests {
				if err := githubClnt.CreateComment(repo, pull, retest); err != nil {
					return err
				}
				glog.Infof("Retesting https://github.com/%s/%s/pull/%d.", owner, repo, *pr.Number)
			} else {
				glog.Infof("Already retested https://github.com/%s/%s/pull/%d %d times. Skipping.", owner, repo, *pr.Number, retestCount)
			}
		}
	}
	return nil
}

func init() {
	flag.Parse()
	u.AssertNotEmpty("owner", owner)
	u.AssertNotEmpty("token_file", tokenFile)
	token, err := u.GetAPITokenFromFile(*tokenFile)
	if err != nil {
		glog.Fatalf("Error accessing user supplied token_file: %v\n", err)
	}
	githubClnt = u.NewGithubClient(*owner, token)
}

func main() {
	u.AssertNotEmpty("repo", repo)

	var err error
	switch *op {
	case "fastForward":
		u.AssertNotEmpty("base_branch", baseBranch)
		u.AssertNotEmpty("ref_sha", refSHA)
		err = fastForward(*repo, *baseBranch, *refSHA)
	// the following three cases are related to release pipeline
	case "newReleaseRequest":
		u.AssertNotEmpty("pipeline", pipelineType)
		u.AssertNotEmpty("tag", tag)
		u.AssertNotEmpty("base_branch", baseBranch)
		u.AssertNotEmpty("ref_sha", refSHA)
		err = CreateReleaseRequest(*repo, *pipelineType, *tag, *baseBranch, *refSHA)
	case "cleanupReleaseRequests":
		err = CleanupReleaseRequests(*owner, *repo)
	case "getBaseSHA":
		var baseSha string
		baseSha, err = getBaseSha(*repo, *prNum)
		if err == nil {
			fmt.Print(baseSha)
		}
	default:
		err = fmt.Errorf("unsupported operation: %s", *op)
	}

	if err != nil {
		glog.Info(err)
		os.Exit(1)
	}
}

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
	releaseOrg   = flag.String("rel_org", "istio-releases", "GitHub Release Org")
	prNum        = flag.Int("pr_num", 0, "PR number")

	githubClnt *u.GithubClient
	ghClntRel  *u.GithubClient
)

const (
	masterBranch = "master"
	pipelineRepo = "pipeline"
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

// CreateReleaseRequest triggers release pipeline by creating a PR.
func CreateReleaseRequest(baseBranch *string) error {
	u.AssertNotEmpty("pipeline", pipelineType)
	u.AssertNotEmpty("tag", tag)
	u.AssertNotEmpty("base_branch", baseBranch)
	u.AssertNotEmpty("ref_sha", refSHA)
	dstBranch := *baseBranch
	glog.Infof("Creating PR to trigger build on %s branch\n", dstBranch)
	prTitle := fmt.Sprintf("%s %s", strings.ToUpper(*pipelineType), *tag)
	prBody := "This is a generated PR that triggers a release, and will be automatically merged when all required tests have passed."
	timestamp := fmt.Sprintf("%v", time.Now().UnixNano())
	srcBranch := "release_" + timestamp
	edit := func() error {
		f, err := os.Create("./" + *pipelineType + "/release_params.sh")
		if err != nil {
			return err
		}
		defer func() {
			cerr := f.Close()
			if err != nil {
				glog.Info(cerr)
			}
		}()

		if _, err := f.WriteString("export CB_BRANCH=" + *baseBranch + "\n"); err != nil {
			return err
		}
		if _, err := f.WriteString("export CB_PIPELINE_TYPE=" + *pipelineType + "\n"); err != nil {
			return err
		}
		if _, err := f.WriteString("export CB_VERSION=" + *tag + "\n"); err != nil {
			return err
		}
		if _, err := f.WriteString("export CB_COMMIT=" + *refSHA + "\n"); err != nil {
			return err
		}
		return nil
	}
	_, err := ghClntRel.CreatePRUpdateRepo(srcBranch, dstBranch, pipelineRepo, prTitle, prBody, edit)
	return err
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
	case "newReleaseRequest":
		if err := CreateReleaseRequest(baseBranch); err != nil {
			glog.Infof("Error during ReleasePipelineBuild: %v\n", err)
			os.Exit(1)
		}
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

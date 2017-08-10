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
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/github"
	multierror "github.com/hashicorp/go-multierror"
	"golang.org/x/oauth2"

	"istio.io/test-infra/toolbox/util"
)

var (
	ci = util.NewCIState()
)

const (
	prTitlePrefix = "[DO NOT MERGE] Auto PR to update dependencies of "
)

type githubClient struct {
	client *github.Client
	owner  string
	token  string
}

func newGithubClient(owner, token string) (*githubClient, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	return &githubClient{client, owner, token}, nil
}

// Generates the url to the remote repository on github
// embedded with proper username and token
func (g githubClient) remote(repo string) string {
	return fmt.Sprintf(
		"https://%s:%s@github.com/%s/%s.git",
		g.owner, g.token, g.owner, repo,
	)
}

// Create a pull request within repo from branch to master
func (g githubClient) createPullRequest(branch, baseBranch, repo string) error {
	if branch == "" {
		return errors.New("branch cannot be empty")
	}
	title := prTitlePrefix + repo
	body := "This PR will be merged automatically once checks are successful."
	req := github.NewPullRequest{
		Head:  &branch,
		Base:  &baseBranch,
		Title: &title,
		Body:  &body,
	}
	log.Printf("Creating a PR with Title: \"%s\" for repo %s", title, repo)
	pr, _, err := g.client.PullRequests.Create(
		context.Background(), g.owner, repo, &req)
	if err != nil {
		return err
	}
	log.Printf("Created new PR at %s", *pr.HTMLURL)
	return nil
}

// Returns a list of repos under the provided owner
func (g githubClient) listRepos() ([]string, error) {
	opt := &github.RepositoryListOptions{Type: "owner"}
	repos, _, err := g.client.Repositories.List(context.Background(), g.owner, opt)
	if err != nil {
		return nil, err
	}
	var listRepoNames []string
	for _, r := range repos {
		listRepoNames = append(listRepoNames, *r.Name)
	}
	return listRepoNames, nil
}

// Checks if a given branch name already exists on remote repo
// Must get a full list of branches and iterate through since
// fetching a nonexisting branch directly results in error
func (g githubClient) existBranch(repo, branch string) (bool, error) {
	branches, _, err := g.client.Repositories.ListBranches(
		context.Background(), g.owner, repo, nil)
	if err != nil {
		return false, err
	}
	for _, b := range branches {
		if b.GetName() == branch {
			return true, nil
		}
	}
	return false, nil
}

// Checks all auto PRs to update dependencies
// Close the ones that have failed the presubmit
// Deletes the remote branches from which the PRs are made
func (g githubClient) closeFailedPullRequests(repo, baseBranch string) error {
	log.Printf("Search for failed auto PRs to update dependencies in repo %s", repo)
	options := github.PullRequestListOptions{
		Base:  baseBranch,
		State: "open",
	}
	prs, _, err := g.client.PullRequests.List(
		context.Background(), g.owner, repo, &options)
	if err != nil {
		return err
	}
	var multiErr error
	for _, pr := range prs {
		if !strings.HasPrefix(*pr.Title, prTitlePrefix) {
			continue
		}
		combinedStatus, _, err := g.client.Repositories.GetCombinedStatus(
			context.Background(), g.owner, repo, *pr.Head.SHA, nil)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
		if util.GetCIState(combinedStatus, nil) == ci.Failure {
			prName := fmt.Sprintf("%s/%s#%d", g.owner, repo, *pr.Number)
			log.Printf("Closing PR %s and deleting branch %s", prName, *pr.Head.Ref)
			*pr.State = "closed"
			if _, _, err := g.client.PullRequests.Edit(
				context.Background(), g.owner, repo, *pr.Number, pr); err != nil {
				multiErr = multierror.Append(multiErr, err)
				log.Printf("Failed to close %s", prName)
			}
			ref := fmt.Sprintf("refs/heads/%s", *pr.Head.Ref)
			if _, err := g.client.Git.DeleteRef(context.Background(), g.owner, repo, ref); err != nil {
				multiErr = multierror.Append(multiErr, err)
				log.Printf("Failed to delete branch %s in repo %s", *pr.Head.Ref, repo)
			}
		}
	}
	return multiErr
}

// Returns the SHA of the commit to which the HEAD of branch points
func (g githubClient) getHeadCommitSHA(repo, branch string) (string, error) {
	ref, _, err := g.client.Git.GetRef(
		context.Background(), g.owner, repo, "refs/heads/"+branch)
	if err != nil {
		return "", err
	}
	return *ref.Object.SHA, nil
}

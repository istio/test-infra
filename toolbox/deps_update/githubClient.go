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

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"istio.io/test-infra/toolbox/util"
)

var (
	ci = util.NewCIState()
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

func (g githubClient) createPullRequest(branch, repo string) error {
	if branch == "" {
		return errors.New("branch cannot be empty")
	}
	title := fmt.Sprintf("[DO NOT MERGE] Auto PR to update dependencies of %s", repo)
	body := "This PR will be merged automatically once checks are successful."
	base := "master"
	req := github.NewPullRequest{
		Head:  &branch,
		Base:  &base,
		Title: &title,
		Body:  &body,
	}
	log.Printf("Creating a PR with Title: \"%s\" for repo %s", title, repo)
	pr, _, err := g.client.PullRequests.Create(context.Background(), g.owner, repo, &req)
	if err != nil {
		return err
	}
	log.Printf("Created new PR at %s", *pr.HTMLURL)
	return nil
}

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

func (g githubClient) getListBranches(repo string) ([]string, error) {
	branches, _, err := g.client.Repositories.ListBranches(
		context.Background(), g.owner, repo, nil)
	if err != nil {
		return nil, err
	}
	var branchNames []string
	for _, b := range branches {
		branchNames = append(branchNames, b.GetName())
	}
	return branchNames, nil
}

func (g githubClient) hasFailedAnyCICheck(repo, branch string) (bool, error) {
	// TODO (chx) list pr, use pr commit sha to get combined status
	// TODO (chx) test with istio token
	combinedStatus, _, err := g.client.Repositories.GetCombinedStatus(
		context.Background(), g.owner, repo, branch, nil)
	if err != nil {
		return false, err
	}
	finalState := util.GetCIState(combinedStatus, nil)
	return (finalState == ci.Failure), nil
}

func (g githubClient) getHeadCommitSHA(repo, branch string) (string, error) {
	ref, _, err := g.client.Git.GetRef(context.Background(), g.owner, repo, "refs/heads/"+branch)
	if err != nil {
		return "", err
	}
	return *ref.Object.SHA, nil
}

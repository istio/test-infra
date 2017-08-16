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

package util

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/github"
	multierror "github.com/hashicorp/go-multierror"
	"golang.org/x/oauth2"
)

// GithubClient masks RPCs to github as local procedures
type GithubClient struct {
	client *github.Client
	owner  string
	token  string
}

// NewGithubClient creates a new GithubClient with proper authentication
func NewGithubClient(owner, token string) (*GithubClient, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	return &GithubClient{client, owner, token}, nil
}

// FastForward moves :branch on :repo to the given sha
func (g GithubClient) FastForward(repo, branch, sha string) error {
	ref := fmt.Sprintf("refs/heads/%s", branch)
	log.Printf("Updating ref %s to commit %s on repo %s", ref, sha, repo)
	refType := "commit"
	gho := github.GitObject{
		SHA:  &sha,
		Type: &refType,
	}
	r := github.Reference{
		Ref:    &ref,
		Object: &gho,
	}
	r.Ref = new(string)
	*r.Ref = ref
	_, _, err := g.client.Git.UpdateRef(
		context.Background(), g.owner, repo, &r, false)
	return err
}

// Remote generates the url to the remote repository on github
// embedded with username and token
func (g GithubClient) Remote(repo string) string {
	return fmt.Sprintf(
		"https://%s:%s@github.com/%s/%s.git",
		g.owner, g.token, g.owner, repo,
	)
}

// CreatePullRequest within :repo from :branch to :baseBranch
func (g GithubClient) CreatePullRequest(
	title, body, branch, baseBranch, repo string) error {
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

// ListRepos returns a list of repos under the provided owner
func (g GithubClient) ListRepos() ([]string, error) {
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

// ExistBranch checks if a given branch name has already existed on remote repo
// Must get a full list of branches and iterate through since
// fetching a nonexisting branch directly results in error
func (g GithubClient) ExistBranch(repo, branch string) (bool, error) {
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

// CloseFailedPullRequests checks all open PRs on baseBranch in repo,
// closes the ones that have failed the presubmit, and deletes the
// remote branches from which the PRs are made
func (g GithubClient) CloseFailedPullRequests(prTitlePrefix, repo, baseBranch string) error {
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
		if GetCIState(combinedStatus, nil) == ci.Failure {
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

// GetHeadCommitSHA finds the SHA of the commit to which the HEAD of branch points
func (g GithubClient) GetHeadCommitSHA(repo, branch string) (string, error) {
	ref, _, err := g.client.Git.GetRef(
		context.Background(), g.owner, repo, "refs/heads/"+branch)
	if err != nil {
		return "", err
	}
	return *ref.Object.SHA, nil
}

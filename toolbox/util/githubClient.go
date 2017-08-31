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
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/github"
	multierror "github.com/hashicorp/go-multierror"
	"golang.org/x/oauth2"
)

var (
	commitType = "commit"
)

const (
	maxCommitDistance = 200
)

// GithubClient masks RPCs to github as local procedures
type GithubClient struct {
	client *github.Client
	owner  string
	token  string
}

// NewGithubClient creates a new GithubClient with proper authentication
func NewGithubClient(owner, token string) *GithubClient {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	return &GithubClient{client, owner, token}
}

// NewGithubClientNoAuth creates a new GithubClient without authentication
// useful when only making GET requests
func NewGithubClientNoAuth(owner string) *GithubClient {
	client := github.NewClient(nil)
	return &GithubClient{client, owner, ""}
}

// SHAIsAncestorOfBranch checks if sha is ancestor of branch
func (g GithubClient) SHAIsAncestorOfBranch(repo, branch, targetSHA string) (bool, error) {
	log.Printf("Checking if %s is ancestor of branch %s", targetSHA, branch)
	sha, err := g.GetHeadCommitSHA(repo, branch)
	if err != nil {
		return false, err
	}
	noEarlierThan, err := g.GetCommitCreationTime(repo, targetSHA)
	if err != nil {
		return false, err
	}
	for i := 0; i < maxCommitDistance; i++ {
		if targetSHA == sha {
			return true, nil
		}
		commit, _, err := g.client.Repositories.GetCommit(
			context.Background(), g.owner, repo, sha)
		if err != nil {
			return false, err
		}
		creationTime, err := g.GetCommitCreationTime(repo, sha)
		if err != nil {
			return false, err
		}
		if creationTime.Before(*noEarlierThan) {
			return false, nil
		}
		sha = *commit.Parents[0].SHA
	}
	err = fmt.Errorf("exceed max iterations %d", maxCommitDistance)
	return false, err
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
	title, body, branch, baseBranch, repo string) (*github.PullRequest, error) {
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
		return nil, err
	}
	log.Printf("Created new PR at %s", *pr.HTMLURL)
	return pr, nil
}

// AddAutoMergeLabelsToPR adds /lgtm and /approve labels to a PR,
// essentially automatically merges the PR without review, if PR passes presubmit
func (g GithubClient) AddAutoMergeLabelsToPR(repo string, pr *github.PullRequest) error {
	return g.AddlabelsToPR(repo, pr, "lgtm", "approved", "release-note-none")
}

// AddlabelsToPR adds labels to the pull request
func (g GithubClient) AddlabelsToPR(
	repo string, pr *github.PullRequest, labels ...string) error {
	_, _, err := g.client.Issues.AddLabelsToIssue(
		context.Background(), g.owner, repo, pr.GetNumber(), labels)
	return err
}

// ClosePRDeleteBranch closes a PR and deletes the branch from which the PR is made
func (g GithubClient) ClosePRDeleteBranch(repo string, pr *github.PullRequest) error {
	prName := fmt.Sprintf("%s/%s#%d", g.owner, repo, pr.GetNumber())
	log.Printf("Closing PR %s and deleting branch %s", prName, *pr.Head.Ref)
	*pr.State = "closed"
	if _, _, err := g.client.PullRequests.Edit(
		context.Background(), g.owner, repo, pr.GetNumber(), pr); err != nil {
		log.Printf("Failed to close %s", prName)
		return err
	}
	ref := fmt.Sprintf("refs/heads/%s", *pr.Head.Ref)
	if _, err := g.client.Git.DeleteRef(context.Background(), g.owner, repo, ref); err != nil {
		log.Printf("Failed to delete branch %s in repo %s", *pr.Head.Ref, repo)
		return err
	}
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

// CloseIdlePullRequests checks all open PRs auto-created on baseBranch in repo,
// closes the ones that have stayed open for a long time, and deletes the
// remote branches from which the PRs are made
func (g GithubClient) CloseIdlePullRequests(prTitlePrefix, repo, baseBranch string) error {
	log.Printf("If any, close failed auto PRs to update dependencies in repo %s", repo)
	idleTimeout := time.Hour * 24
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
		if time.Since(*pr.CreatedAt).Nanoseconds() > idleTimeout.Nanoseconds() {
			if err := g.ClosePRDeleteBranch(repo, pr); err != nil {
				multiErr = multierror.Append(multiErr, err)
			}
		}
	}
	return multiErr
}

// GetHeadCommitSHA finds the SHA of the commit to which the HEAD of branch points
func (g GithubClient) GetHeadCommitSHA(repo, branch string) (string, error) {
	return g.getReferenceSHA(repo, "refs/heads/"+branch)
}

// GetTagCommitSHA finds the SHA of the commit from which the tag was made
func (g GithubClient) GetTagCommitSHA(repo, tag string) (string, error) {
	sha, err := g.getReferenceSHA(repo, "refs/tags/"+tag)
	if err != nil {
		return "", err
	}
	tagObj, _, err := g.client.Git.GetTag(
		context.Background(), g.owner, repo, sha)
	if err != nil {
		return "", err
	}
	return *tagObj.Object.SHA, nil
}

// GetCommitCreationTime gets the time when the commit identified by sha is created
func (g GithubClient) GetCommitCreationTime(repo, sha string) (*time.Time, error) {
	commit, _, err := g.client.Git.GetCommit(
		context.Background(), g.owner, repo, sha)
	if err != nil {
		return nil, err
	}
	return (*(*commit).Author).Date, nil
}

// GetCommitCreationTimeByTag finds the time when the commit pointed by a tag is created
// Note that SHA of the tag is different from the commit SHA
func (g GithubClient) GetCommitCreationTimeByTag(repo, tag string) (*time.Time, error) {
	commitSHA, err := g.GetTagCommitSHA(repo, tag)
	if err != nil {
		return nil, err
	}
	return g.GetCommitCreationTime(repo, commitSHA)
}

// GetFileContent retrieves the file content from the hosted repo
func (g GithubClient) GetFileContent(repo, branch, path string) (string, error) {
	opt := github.RepositoryContentGetOptions{branch}
	fileContent, _, _, err := g.client.Repositories.GetContents(
		context.Background(), g.owner, repo, path, &opt)
	if err != nil {
		return "", err
	}
	return fileContent.GetContent()
}

// CreateAnnotatedTag retrieves the file content from the hosted repo
func (g GithubClient) CreateAnnotatedTag(repo, tag, sha, msg string) error {
	if !SHARegex.MatchString(sha) {
		return fmt.Errorf(
			"unable to create tag %s on repo %s: invalid commit SHA %s",
			tag, repo, sha)
	}
	tagObj := github.Tag{
		Tag:     &tag,
		Message: &msg,
		Object: &github.GitObject{
			Type: &commitType,
			SHA:  &sha,
		},
	}
	tagResponse, _, err := g.client.Git.CreateTag(
		context.Background(), g.owner, repo, &tagObj)
	if err != nil {
		log.Printf("Failed to create tag %s on repo %s", tag, repo)
		return err
	}
	refString := "refs/tags/" + tag
	refObj := github.Reference{
		Ref: &refString,
		Object: &github.GitObject{
			Type: &commitType,
			SHA:  tagResponse.SHA,
		},
	}
	_, _, err = g.client.Git.CreateRef(
		context.Background(), g.owner, repo, &refObj)
	if err != nil {
		log.Printf("Failed to create reference with tag just created: %s", tag)
	}
	return err
}

// CreateReleaseUploadArchives creates a release given release tag and
// upload all files in archiveDir as assets of this release
func (g GithubClient) CreateReleaseUploadArchives(repo, releaseTag, archiveDir string) error {
	// create release
	release := github.RepositoryRelease{TagName: &releaseTag}
	log.Printf("Creating on github new %s release [%s]\n", repo, releaseTag)
	res, _, err := g.client.Repositories.CreateRelease(
		context.Background(), g.owner, repo, &release)
	if err != nil {
		log.Printf("Failed to create new release on repo %s with releaseTag: %s", repo, releaseTag)
		return err
	}
	releaseID := *res.ID
	// upload archives
	files, err := ioutil.ReadDir(archiveDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		log.Printf("Uploading asset %s\n", f.Name())
		filePath := fmt.Sprintf("%s/%s", archiveDir, f.Name())
		fd, err := os.Open(filePath)
		if err != nil {
			return err
		}
		opt := github.UploadOptions{f.Name()}
		_, _, err = g.client.Repositories.UploadReleaseAsset(
			context.Background(), g.owner, repo, releaseID, &opt, fd)
		if err != nil {
			log.Printf("Failed to upload asset %s to release %s on repo %s: %s",
				f.Name(), releaseTag, repo)
			return err
		}
	}
	return nil
}

func (g GithubClient) getReferenceSHA(repo, ref string) (string, error) {
	githubRefObj, _, err := g.client.Git.GetRef(
		context.Background(), g.owner, repo, ref)
	if err != nil {
		return "", err
	}
	return *githubRefObj.Object.SHA, nil
}

// SearchIssues get issues/prs based on query
func (g GithubClient) SearchIssues(queries []string, keyWord, sort, order string) (*github.IssuesSearchResult, error) {
	q := strings.Join(queries, " ")

	searchOption := &github.SearchOptions{
		Sort:  sort,
		Order: order,
	}

	issueResult, _, err := g.client.Search.Issues(context.Background(), q, searchOption)
	if err != nil {
		log.Printf("Failed to search issues")
		return nil, err
	}

	return issueResult, nil
}

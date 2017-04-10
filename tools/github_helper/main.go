package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	tokenFile   = flag.String("token_file", "", "File containing Auth Token.")
	owner       = flag.String("owner", "istio", "Github Owner or org.")
	repos       = flag.String("repos", "", "Comma separated list of Github repo within the org.")
	base        = flag.String("base", "stable", "The base branch used for PR.")
	head        = flag.String("head", "master", "The head branch used for PR.")
	pullRequest = flag.Int("pr", 0, "The Pull request to use.")
	checkToSkip = flag.String("check_to_skip", "", "Lists of check(s) can be skipped, full context separated with comma.")
	fastForward = flag.Bool("fast_forward", false, "Creates a PR updating Base to Head.")
	verify      = flag.Bool("verify", false, "Verifies PR on Base and push them if success.")
	comment     = flag.String("comment", "", "The comment to send to the Pull Request.")
	GH          = newGhConst()
)

type ghConst struct {
	success string
	failure string
	pending string
	closed  string
	all     string
	commit  string
}

// Simple Github Helper
type helper struct {
	Owner   string
	Repo    string
	Base    string
	Head    string
	Pr      int
	CheckToSkip  []string
	Client  *github.Client
}

// Get token from tokenFile is set, otherwise is anonymous.
func getToken() (*http.Client, error) {
	if *tokenFile == "" {
		return nil, nil
	}
	b, err := ioutil.ReadFile(*tokenFile)
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(string(b[:]))
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(token[:])})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	return tc, nil
}

// Creates a new ghConst
func newGhConst() *ghConst {
	return &ghConst{
		success: "success",
		failure: "failure",
		pending: "pending",
		closed:  "closed",
		all:     "all",
		commit:  "commit",
	}
}

// Creates a new Github Helper from provided
func newHelper(r *string) (*helper, error) {
	if tc, err := getToken(); err == nil {
		if *repos == "" {
			return nil, errors.New("repo flag must be set!")
		}
		client := github.NewClient(tc)
		return &helper{
			Owner:   *owner,
			Repo:    *r,
			Base:    *base,
			Head:    *head,
			Pr:      *pullRequest,
			CheckToSkip:  strings.Split(*checkToSkip, ","),
			Client:  client,
		}, nil
	} else {
		return nil, err
	}
}

// Create a new branch for fast forward
func (h helper) createBranchForFastForward(commit *string) (*string, error) {
	branchName := fmt.Sprintf("fastForward-%s-%s", h.Head, (*commit)[0:7])
	ref := fmt.Sprintf("refs/heads/%s", branchName)

	gho := github.GitObject{
		SHA: commit,
	}
	r := github.Reference{
		Ref:    &ref,
		Object: &gho,
	}

	if _, resp, err := h.Client.Git.CreateRef(context.TODO(), h.Owner, h.Repo, &r); err == nil {
		log.Printf("Created a new branch for fast forward: %s", branchName)
		return &branchName, nil
	} else {
		if resp.StatusCode == 422 {
			log.Printf("Branch %s already exists in repo %s!", branchName, h.Repo)
		}
		return nil, err
	}
}

// Creates a pull request from Base branch
func (h helper) createPullRequestToBase(commit *string) error {
	if commit == nil {
		return errors.New("commit cannot be nil.")
	}

	newHead, err := h.createBranchForFastForward(commit)
	if err != nil {
		return err
	}

	title := fmt.Sprintf(
		"DO NOT MERGE! Fast Forward %s to %s.", h.Base, *commit)
	body := "This PR will be merged automatically once checks are successful."
	req := github.NewPullRequest{
		Head:  newHead,
		Base:  &h.Base,
		Title: &title,
		Body:  &body,
	}
	log.Printf("Creating a PR with Title: \"%s\" for repo %s", title, h.Repo)
	if pr, _, err := h.Client.PullRequests.Create(context.TODO(), h.Owner, h.Repo, &req); err == nil {
		log.Printf("Created new PR at %s", *pr.HTMLURL)
	} else {
		return err
	}
	return nil
}

// Gets the last commit from Head branch.
func (h helper) getLastCommitFromHead() (*string, error) {
	comp, _, err := h.Client.Repositories.CompareCommits(context.TODO(), h.Owner, h.Repo, h.Head, h.Base)
	if err == nil {
		if *comp.BehindBy > 0 {
			commit := comp.BaseCommit.SHA
			log.Printf(
				"%s is %d commits ahead from %s, and HEAD commit is %s in repo %s",
				h.Head, *comp.BehindBy, h.Base, *commit, h.Repo)
			return commit, nil
		}
	}
	return nil, err
}

// Fast forward Base branch to the last commit of Head branch.
func (h helper) fastForwardBase() error {
	commit, err := h.getLastCommitFromHead()
	if err != nil {
		return err
	}
	if commit != nil {
		options := github.PullRequestListOptions{
			Head:  h.Head,
			Base:  h.Base,
			State: GH.all,
		}

		prs, _, err := h.Client.PullRequests.List(context.TODO(), h.Owner, h.Repo, &options)
		if err == nil {
			for _, pr := range prs {
				if strings.Contains(*pr.Title, *commit) {
					log.Printf("A PR already exist for %s on repo %s", *commit, h.Repo)
					return nil
				}
			}
		}
		return h.createPullRequestToBase(commit)
	}
	log.Printf("Branches %s and %s are in sync for repo %s.", h.Base, h.Head, h.Repo)
	return nil
}

// Close an existing PR
func (h helper) closePullRequest(pr *github.PullRequest) error {
	log.Printf("Closing PR %d on repo %s", *pr.Number, h.Repo)
	*pr.State = GH.closed
	_, _, err := h.Client.PullRequests.Edit(context.TODO(), h.Owner, h.Repo, *pr.Number, pr)
	return err
}

// Create an annotated stable tag from the given commit.
func (h helper) createStableTag(commit *string) error {
	if commit == nil {
		return errors.New("commit cannot be nil.")
	}
	sha := *commit
	tag := fmt.Sprintf("stable-%s", sha[0:7])
	message := "Stable build"
	log.Printf("Creating tag %s on %s for commit %s in repo %s", tag, h.Base, *commit, h.Repo)
	gho := github.GitObject{
		SHA:  commit,
		Type: &GH.commit,
	}
	gt := github.Tag{
		Object:  &gho,
		Message: &message,
		Tag:     &tag,
	}
	t, resp, err := h.Client.Git.CreateTag(context.TODO(), h.Owner, h.Repo, &gt)
	if err != nil {
		return err
	}
	log.Printf("Creating ref tag %s on %s for commit %s in repo %s", tag, h.Base, *commit, h.Repo)
	ref := fmt.Sprintf("refs/tags/%s", tag)
	// Getting the SHA from the annotated tag
	at := github.GitObject{
		SHA:  t.SHA,
		Type: &GH.commit,
	}
	r := github.Reference{
		Ref:    &ref,
		Object: &at,
	}
	_, resp, err = h.Client.Git.CreateRef(context.TODO(), h.Owner, h.Repo, &r)
	// Already exists
	if resp.StatusCode != 422 {
		return err
	}
	return nil
}

// Update the Base branch reference to a given commit.
func (h helper) updateBaseReference(commit *string) error {
	if commit == nil {
		return errors.New("commit cannot be nil")
	}
	ref := fmt.Sprintf("refs/heads/%s", h.Base)
	log.Printf("Updating ref %s to commit %s for repo %s", ref, *commit, h.Repo)
	gho := github.GitObject{
		SHA:  commit,
		Type: &GH.commit,
	}
	r := github.Reference{
		Ref:    &ref,
		Object: &gho,
	}
	r.Ref = new(string)
	*r.Ref = ref

	_, _, err := h.Client.Git.UpdateRef(context.TODO(), h.Owner, h.Repo, &r, false)
	return err
}

// Deletes the new branch created for fast forward
func (h helper) deleteFastForwardBranch(head string) {
	ref := fmt.Sprintf("refs/heads/%s", head)
	if _, err := h.Client.Git.DeleteRef(context.TODO(), h.Owner, h.Repo, ref); err != nil {
		log.Panicf("Failed to delete fast forward branch %s in repo %s", head, h.Repo)
	}
}

// Checks if a PR is ready to be pushed. Create a stable tag and
// fast forward Base to the PR's head commit.
func (h helper) updatePullRequest(pr *github.PullRequest, s *github.CombinedStatus) error {
	state := *s.State
	if *s.TotalCount == 0 {
		// There seems to be a bug in the Github API when no check are set.
		// The status stays in pending state forever
		state = GH.success
	}

	if state == GH.failure && (len(h.CheckToSkip) > 0) {
		privilege := true
		for _, status := range s.Statuses {
			if *status.State != GH.success {
				skip := false
				for _, check := range h.CheckToSkip {
					if *status.Context == check {
						//Find a match so that this failure can be skipped
						skip = true
						break
					}
				}
				if !skip {
					//If this failure cannot be skipped, this fast-forward is not eligible for privilege
					privilege = false
					break
				}
			}
		}
		if privilege {
			state = GH.success
		}
	}

	switch state {
	case GH.success:
		if err := h.createStableTag(s.SHA); err == nil {
			if err := h.updateBaseReference(s.SHA); err != nil {
				log.Printf("Could not update %s reference to %s for repo %s.\n%v", h.Base, *s.SHA, h.Repo, err)
				return nil
			}
			h.deleteFastForwardBranch(*pr.Head.Ref)
			// Note there is no need to close the PR here.
			// It will be done automatically once Base ref is updated
		} else {
			// We may fix github problems and manually merge, so keep the branch.
			return err
		}
	case GH.failure:
		h.deleteFastForwardBranch(*pr.Head.Ref)
		return h.closePullRequest(pr)
	case GH.pending:
		log.Printf("Pull Request %d is still being checked for repo %s", *pr.Number, h.Repo)
	}
	return nil
}

// Checks all the PR on Base and calls updatePullRequest on each.
func (h helper) verifyPullRequestStatus() error {
	options := github.PullRequestListOptions{
		Base:  h.Base,
		State: "open",
	}
	prs, _, err := h.Client.PullRequests.List(context.TODO(), h.Owner, h.Repo, &options)
	if err != nil {
		return err
	}
	for _, pr := range prs {
		if !strings.Contains(*pr.Title, "DO NOT MERGE! Fast Forward") {
			continue
		}
		statuses, _, err := h.Client.Repositories.GetCombinedStatus(
			context.TODO(), h.Owner, h.Repo, *pr.Head.SHA, new(github.ListOptions))
		if err == nil {
			err = h.updatePullRequest(pr, statuses)
		}
		if err != nil {
			log.Fatalf("Could not update PR %d for repo %s. \n%v", *pr.Number, h.Repo, err)
		}
	}
	log.Printf("No more PR to verify for branch %s in repo %s.", h.Base, h.Repo)
	return nil
}

// Creates a comment on a Pull Request
func (h helper) createComment(comment *string) error {
	if h.Pr <= 0 {
		return errors.New("PR number needs to be greather than 0")
	}
	c := github.IssueComment{
		Body: comment,
	}
	log.Printf("Commenting \"%s\" on PR %d for %s/%s", *comment, h.Pr, h.Owner, h.Repo)
	_, _, err := h.Client.Issues.CreateComment(context.TODO(), h.Owner, h.Repo, h.Pr, &c)
	return err
}

func main() {
	flag.Parse()
	if *verify {
		for _, r := range strings.Split(*repos, ",") {
			h, err := newHelper(&r)
			if err != nil {
				log.Fatalf("Could not instantiate a github client %v", err)
			}
			if err = h.verifyPullRequestStatus(); err != nil {
				log.Fatalf("Unable to verify PR from %s.\n%v", h.Base, err)
			}
		}
	}
	if *fastForward {
		for _, r := range strings.Split(*repos, ",") {
			h, err := newHelper(&r)
			if err != nil {
				log.Fatalf("Could not instantiate a github client %v", err)
			}
			if err = h.fastForwardBase(); err != nil {
				log.Fatalf("Unable to fast forward %s.\n%v", h.Base, err)
			}
		}
	}
	if *comment != "" {
		for _, r := range strings.Split(*repos, ",") {
			h, err := newHelper(&r)
			if err != nil {
				log.Fatalf("Could not instantiate a github client %v", err)
			}
			if err := h.createComment(comment); err != nil {
				log.Fatalf("Unable to create a comment on PR %d.\n%v", h.Pr, err)
			}
		}
	}
}


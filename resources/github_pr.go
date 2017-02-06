package main

import (
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
	repo        = flag.String("repo", "", "Github repo within the org.")
	base        = flag.String("base", "stable", "The base branch used for PR.")
	head        = flag.String("head", "master", "The head branch used for PR.")
	fastForward = flag.Bool("fast_forward", false, "Creates a PR updating Base to Head.")
	verify      = flag.Bool("verify", false, "Verifies PR on Base and push them if success.")
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
	Owner  string
	Repo   string
	Base   string
	Head   string
	Client *github.Client
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
		closed: "closed",
		all: "all",
		commit: "commit",
	}
}

// Creates a new Github Helper from provided
func newHelper() (*helper, error) {
	if tc, err := getToken(); err == nil {
		if *repo == "" {
			return nil, errors.New("repo flag must be set!")
		}
		client := github.NewClient(tc)
		return &helper{
			Owner:  *owner,
			Repo:   *repo,
			Base:   *base,
			Head:   *head,
			Client: client,
		}, nil
	} else {
		return nil, err
	}
}

// Creates a pull request from Base branch
func (h helper) createPullRequestToBase(commit *string) error {
	if commit == nil {
		return errors.New("commit cannot be nil.")
	}
	title := fmt.Sprintf(
		"Fast Forward %s to %s\n Do not use the UI to merge this PR.", h.Base, *commit)
	req := github.NewPullRequest{
		Head:  &h.Head,
		Base:  &h.Base,
		Title: &title,
	}
	log.Printf("Creating a PR with Title: \"%s\"", title)
	if pr, _, err := h.Client.PullRequests.Create(h.Owner, h.Repo, &req); err == nil {
		log.Printf("Created new PR at %s", *pr.HTMLURL)
	} else {
		return err
	}
	return nil
}

// Gets the last commit from Head branch.
func (h helper) getLastCommitFromHead() (*string, error) {
	comp, _, err := h.Client.Repositories.CompareCommits(h.Owner, h.Repo, h.Head, h.Base)
	if err == nil {
		if *comp.BehindBy > 0 {
			commit := comp.BaseCommit.SHA
			log.Printf(
				"%s is %d commits ahead from %s, and HEAD commit is %s",
				h.Head, *comp.BehindBy, h.Base, *commit)
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

		prs, _, err := h.Client.PullRequests.List(h.Owner, h.Repo, &options)
		if err == nil {
			for _, pr := range prs {
				if strings.Contains(*pr.Title, *commit) {
					log.Printf("A PR already exist for %s", *commit)
					return nil
				}
			}
		}
		return h.createPullRequestToBase(commit)
	}
	log.Printf("Branches %s and %s as in sync.", h.Base, h.Head)
	return nil
}

// Close an existing PR
func (h helper) closePullRequest(pr *github.PullRequest) error {
	log.Printf("Closing PR %d", *pr.ID)
	*pr.State = GH.closed
	_, _, err := h.Client.PullRequests.Edit(h.Owner, h.Repo, *pr.ID, pr)
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
	log.Printf("Creating tag %s on %s for commit %s", tag, h.Base, *commit)
	gho := github.GitObject{
		SHA:  commit,
		Type: &GH.commit,
	}
	gt := github.Tag{
		Object:  &gho,
		Message: &message,
		Tag:     &tag,
	}
	t, resp, err := h.Client.Git.CreateTag(h.Owner, h.Repo, &gt)
	if err != nil {
		return err
	}
	log.Printf("Creating ref tag %s on %s for commit %s", tag, h.Base, *commit)
	ref := fmt.Sprintf("refs/tags/%s", tag)
	r := github.Reference{
		Ref:    &ref,
		Object: t.Object,
	}
	_, resp, err = h.Client.Git.CreateRef(h.Owner, h.Repo, &r)
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
	log.Printf("Updating ref %s to commit %s", ref, *commit)
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

	_, _, err := h.Client.Git.UpdateRef(h.Owner, h.Repo, &r, false)
	return err
}

// Checks if a PR is ready to be pushed. Create a stable tag and
// fast forward Base to the PR's head commit.
func (h helper) updatePullRequest(pr *github.PullRequest, s *github.CombinedStatus) error {
	switch *s.State {
	case GH.success:
		if err := h.createStableTag(s.SHA); err == nil {
			if err := h.updateBaseReference(s.SHA); err != nil {
				log.Printf("Could not update %s reference to %s.\n%v", h.Base, *s.SHA, err)
				return nil
			}
			// Note there is no need to close the PR here.
			// It will be done automatically once Base ref is updated
		} else {
			return err
		}
	case GH.failure:
		log.Printf("Closing PR %d", *pr.ID)
		return h.closePullRequest(pr)
	case GH.pending:
		log.Printf("Pull Request %d is still being checked", pr.ID)
	}
	return nil
}

// Checks all the PR on Base and calls updatePullRequest on each.
func (h helper) verifyPullRequestStatus() error {
	options := github.PullRequestListOptions{
		Head:  h.Head,
		Base:  h.Base,
		State: "open",
	}
	prs, _, err := h.Client.PullRequests.List(h.Owner, h.Repo, &options)
	if err != nil {
		return err
	}
	for _, pr := range prs {
		statuses, _, err := h.Client.Repositories.GetCombinedStatus(
			h.Owner, h.Repo, *pr.Head.SHA, new(github.ListOptions))
		if err == nil {
			err = h.updatePullRequest(pr, statuses)
		}
		if err != nil {
			log.Fatalf("Could not update PR %d. \n%v", *pr.ID, err)
		}
	}
	log.Printf("No more PR to verify for branch %s.", h.Base)
	return nil
}

func main() {
	flag.Parse()
	h, err := newHelper()
	if err != nil {
		log.Fatalf("Could not instantiate a github client %v", err)
	}
	if *verify {
		if err = h.verifyPullRequestStatus(); err != nil {
			log.Fatalf("Unable to verify PR from %s.\n%v", h.Base, err)
		}
	}
	if *fastForward {
		if err = h.fastForwardBase(); err != nil {
			log.Fatalf("Unable to fast forward %s.\n%v", h.Base, err)
		}
	}
}

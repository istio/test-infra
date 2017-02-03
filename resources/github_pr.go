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
)

const (
	SUCCESS string = "success"
	FAILURE string = "failure"
	PENDING string = "pending"
	CLOSED  string = "closed"
	ALL     string = "all"
	COMMIT  string = "commit"
)

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

// Creates a new Github Helper from provided
func newHelper() (*helper, error) {
	if tc, err := getToken(); err == nil {
		client := github.NewClient(tc)
		h := new(helper)
		if *repo == "" {
			return nil, errors.New("repo flag must be set!")
		}
		h.Owner = *owner
		h.Repo = *repo
		h.Base = *base
		h.Head = *head
		h.Client = client
		return h, nil
	} else {
		return nil, err
	}
}

// Creates a pull request from Base branch
func (h helper) createPullRequestToBase(commit *string) error {
	if commit == nil {
		return errors.New("commit cannot be nil.")
	}
	title := fmt.Sprintf("Fast Forward %s to %s", h.Base, *commit)
	req := new(github.NewPullRequest)
	*req.Head = h.Head
	*req.Base = h.Base
	*req.Title = title
	log.Printf("Creating a PR with Title: \"%s\"", title)
	if pr, _, err := h.Client.PullRequests.Create(h.Owner, h.Repo, req); err == nil {
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
		options := new(github.PullRequestListOptions)
		options.Head = h.Head
		options.Base = h.Base
		options.State = ALL

		prs, _, err := h.Client.PullRequests.List(h.Owner, h.Repo, options)
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
	*pr.State = CLOSED
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
	log.Printf("Creating tag %s on %s for commit %s", tag, h.Base, *commit)
	t := new(github.Tag)
	t.Object = new(github.GitObject)
	t.Object.SHA = commit
	t.Object.Type = new(string)
	*t.Object.Type = COMMIT
	t.Message = new(string)
	*t.Message = "Stable build"
	t.Tag = new(string)
	*t.Tag = tag
	t, resp, err := h.Client.Git.CreateTag(h.Owner, h.Repo, t)
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
	r := new(github.Reference)
	r.Ref = new(string)
	*r.Ref = ref
	r.Object = new(github.GitObject)
	r.Object.SHA = commit
	r.Object.Type = new(string)
	*r.Object.Type = COMMIT
	_, _, err := h.Client.Git.UpdateRef(h.Owner, h.Repo, r, false)
	return err
}

// Checks if a PR is ready to be pushed. Create a stable tag and
// fast forward Base to the PR's head commit.
func (h helper) updatePullRequest(pr *github.PullRequest, s *github.CombinedStatus) error {
	switch *s.State {
	case SUCCESS:
		if err := h.createStableTag(s.SHA); err == nil {
			if err := h.updateBaseReference(s.SHA); err != nil {
				log.Printf("Could not update %s reference to %s.\n%v", h.Base, *s.SHA, err)
				return nil
			} else {
				return h.closePullRequest(pr)
			}
		} else {
			return err
		}
	case FAILURE:
		return h.closePullRequest(pr)
	case PENDING:
		log.Printf("Pull Request %d is still being checked", pr.ID)
	}
	return nil
}

// Checks all the PR on Base and calls updatePullRequest on each.
func (h helper) verifyPullRequestStatus() error {
	options := new(github.PullRequestListOptions)
	options.Head = h.Head
	options.Base = h.Base
	options.State = "open"
	prs, _, err := h.Client.PullRequests.List(h.Owner, h.Repo, options)
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
	if *fastForward {
		if err = h.fastForwardBase(); err != nil {
			log.Fatalf("Unable to fast forward %s.\n%v", h.Base, err)
		}
	}
	if *verify {
		if err = h.verifyPullRequestStatus(); err != nil {
			log.Fatalf("Unable to verify PR from %s.\n%v", h.Base, err)
		}
	}
}

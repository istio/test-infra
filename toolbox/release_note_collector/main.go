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
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/github"

	u "istio.io/test-infra/toolbox/util"
)

type (
	orderV string
	sortV  string
)

const (
	// Allowed field values can be found at https://developer.github.com/v3/search/#search-repositories
	desc    = orderV("desc")
	created = sortV("created")
)

var (
	org             = flag.String("user", "istio", "Github owner or org")
	tokenFile       = flag.String("token_file", "", "Github token file (optional)")
	repos           = flag.String("repos", "", "Github repos, separate using \",\"")
	label           = flag.String("label", "release-note", "Release-note label")
	sort            = flag.String("sort", string(created), "The sort field. Can be comments, created, or updated.")
	order           = flag.String("order", string(desc), "The sort order if sort parameter is provided. One of asc or desc.")
	outputFile      = flag.String("output", "./release-note", "Path to output file")
	previousRelease = flag.String("previous_release", "", "Previous release")
	currentRelease  = flag.String("current_release", "", "Current release")
	prLink          = flag.Bool("pr_link", false, "Weather a link of the PR is added at the end of each release note")
	branch          = flag.String("branch", "master", "Commit branch, master or release branch")
	gh              *u.GithubClient
)

func init() {
	flag.Parse()
	u.AssertNotEmpty("previous_release", previousRelease)

	if *tokenFile != "" {
		token, err := u.GetAPITokenFromFile(*tokenFile)
		if err != nil {
			log.Fatalf("Error accessing user supplied token_file: %v\n", err)
		}
		gh = u.NewGithubClient(*org, token)
	} else {
		gh = u.NewGithubClientNoAuth(*org)
	}
}

func main() {
	f, err := os.OpenFile(*outputFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		log.Printf("Failed to open and/or create output file %s", *outputFile)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error during closing file %s: %s\n", *outputFile, err)
		}
	}()

	repoList := strings.Split(*repos, ",")
	for _, repo := range repoList {
		log.Printf("Start fetching release note from %s", repo)
		queries, err := createQueryString(repo)
		if err != nil {
			log.Printf("Failed to create query string for %s", repo)
			continue
		}
		log.Printf("Query: %v", queries)

		issues, err := gh.SearchIssues(queries, *sort, *order)
		if err != nil {
			log.Printf("Failed to fetch PR with release note for %s: %s", repo, err)
			continue
		}
		fetchRelaseNoteFromRepo(repo, issues, f)
	}
}

func fetchRelaseNoteFromRepo(repo string, issues []*github.Issue, f io.StringWriter) {
	title := fmt.Sprintf("\nistio/%s: %s -- %s\n", repo, *previousRelease, *currentRelease)
	if _, err := f.WriteString(title); err != nil {
		log.Printf("Failed to write title into output file: %s. Err: %s", title, err)
	}

	for _, issue := range issues {
		note := fetchReleaseNoteFromPR(issue)
		if note == "" {
			continue
		}
		if *prLink {
			note += fmt.Sprintf("  %s\n", issue.GetHTMLURL())
		}
		if _, err := f.WriteString("* " + note); err != nil {
			log.Printf("Failed to write a note into output file: %s. Err: %s", note, err)
		}
	}
}

func fetchReleaseNoteFromPR(issue *github.Issue) (note string) {
	reg := regexp.MustCompile("```release-note\r\n((?s).+)\r\n```")
	m := reg.FindStringSubmatch(*issue.Body)
	if len(m) == 2 {
		note = m[1]
	}
	note = strings.TrimSpace(note)
	if strings.EqualFold(note, u.ReleaseNoteNone) {
		return ""
	}
	return note
}

func createQueryString(repo string) ([]string, error) {
	var queries []string

	startTime, err := getReleaseTime(repo, *previousRelease)
	if err != nil {
		log.Printf("Failed to get created time of previous release -- %s: %s", *previousRelease, err)
		return nil, err
	}

	if *currentRelease == "" {
		if *currentRelease, err = gh.GetLatestRelease(repo); err != nil {
			log.Printf("Failed to get latest release version when current_release is missing: %s", err)
			return nil, err
		}
		log.Printf("Last release version: %s", *currentRelease)
	}
	endTime, err := getReleaseTime(repo, *currentRelease)
	if err != nil {
		log.Printf("Failed to get created time of current release -- %s: %s", *currentRelease, err)
		return nil, err
	}

	queries = addQuery(queries, "repo", *org, "/", repo)
	queries = addQuery(queries, "label", *label)
	queries = addQuery(queries, "is", "merged")
	queries = addQuery(queries, "type", "pr")
	queries = addQuery(queries, "merged", startTime, "..", endTime)
	queries = addQuery(queries, "base", *branch)

	return queries, nil
}

func addQuery(queries []string, queryParts ...string) []string {
	if len(queryParts) < 2 {
		log.Printf("Not enough to form a query: %v", queryParts)
		return queries
	}
	for _, part := range queryParts {
		if part == "" {
			return queries
		}
	}

	return append(queries, fmt.Sprintf("%s:%s", queryParts[0], strings.Join(queryParts[1:], "")))
}

func getReleaseTime(repo, release string) (string, error) {
	createTime, err := getReleaseTagCreationTime(repo, release)
	if err != nil {
		log.Println("Failed to get created createTime of this release tag")
		return "", err
	}
	t := createTime.UTC()
	timeString := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	return timeString, nil
}

// getReleaseCreationTime gets the creation time of a release tag (0.1.6, 0.2.7)
func getReleaseTagCreationTime(repo, tag string) (createTime time.Time, err error) {
	if repo == "istio" {
		createTime, err = gh.GetReleaseTagCreationTime(repo, tag)
	} else {
		createTime, err = gh.GetannotatedTagCreationTime(repo, tag)
	}
	if err != nil {
		log.Printf("Cannot get the creation time of %s/%s", repo, tag)
		return time.Time{}, err
	}
	return createTime, nil
}

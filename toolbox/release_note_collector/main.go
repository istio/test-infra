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
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
	u "istio.io/test-infra/toolbox/util"
)

const (
	releaseNoteSuffix = ".releasenote"
)

var (
	org       = flag.String("user", "istio", "Github owner or org")
	repos     = flag.String("repos", "", "Github repos, separate using \",\"")
	label     = flag.String("label", "release-note", "Release-note label")
	sort      = flag.String("sort", "create", "The sort field. Can be comments, created, or updated.")
	order     = flag.String("order", "desc", "The sort order if sort parameter is provided. One of asc or desc.")
	output    = flag.String("output", "./", "Path to output file")
	previousRelease = flag.String("previous_release", "", "Previous release")
	currentRelease   = flag.String("current_release", "", "Current release")
	gh *u.GithubClient
)

func main() {
	flag.Parse()
	if *previousRelease == "" {
		log.Printf("Error: You need to specfy a previous release")
		os.Exit(1)
	}
	gh = u.NewGithubClientNoAuth(*org)

	repoList := strings.Split(*repos, ",")
	for _, repo := range repoList {
		log.Printf("Start fetching release note from %s", repo)
		queries, err := createQueryString(repo)
		if err != nil {
			log.Printf("Failed to create query string for %s", repo)
			continue
		}
		log.Printf("Query: %v", queries)

		issuesResult, err := gh.SearchIssues(queries, "", *sort, *order)
		if err != nil {
			log.Printf("Failed to fetch PR with release note for %s: %s", repo, err)
			continue
		}
		if err = fetchRelaseNoteFromRepo(repo, issuesResult); err != nil {
			log.Printf("Failed to get release note for %s: %s", repo, err)
			continue
		}
	}

}

func fetchRelaseNoteFromRepo(repo string, issuesResult *github.IssuesSearchResult) error {
	fileName := filepath.Join(*output, repo+releaseNoteSuffix)
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		log.Printf("Failed to create output file %s", fileName)
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error during closing file %s: %s\n", fileName, err)
		}
	}()

	f.WriteString(fmt.Sprintf("%s: %s -- %s release note\n", *org, repo, *currentRelease))
	f.WriteString(fmt.Sprintf("Previous release version: %s", *previousRelease))
	f.WriteString(fmt.Sprintf("Total: %d\n", *issuesResult.Total))
	for _, i := range issuesResult.Issues {
		note := fetchReleaseNoteFromPR(i)
		f.WriteString(note)
	}
	if issuesResult.GetIncompleteResults() {
		f.WriteString("!!Warning: Some release notes missing due to incomplete search result from github!!")
	}
	return nil
}

func fetchReleaseNoteFromPR(i github.Issue) (note string) {
	reg := regexp.MustCompile("```release-note((?s).*)```")
	m := reg.FindStringSubmatch(*i.Body)
	if len(m) == 2 {
		note = m[1]
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
	time, err := gh.GetCommitCreationTimeByTag(repo, release)
	if err != nil {
		log.Println("Failed to get created time of this release tag")
		return "", err
	}
	t := time.UTC()
	log.Printf("Format time: %s", t.Format("2017-07-08T07:13:35Z"))
	timeString := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	return timeString, nil
}

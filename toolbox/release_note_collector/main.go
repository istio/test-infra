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
	version   = flag.String("version", "", "Release version")
	output    = flag.String("output", "./", "Path to output file")
	startDate = flag.String("start_date", "", "Start date")
	endDate   = flag.String("end_date", "", "End date")
)

func main() {
	flag.Parse()

	repoList := strings.Split(*repos, ",")
	for _, repo := range repoList {
		log.Printf("Start fetching release note from %s", repo)
		queries := createQueryString(repo)
		log.Printf("Query: %s", queries)

		g := u.NewGithubClientNoAuth(*org)
		issuesResult, err := g.SearchIssues(queries, "", *sort, *order)
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

	f.WriteString(fmt.Sprintf("%s: %s -- %s release note\n", *org, repo, *version))
	f.WriteString(fmt.Sprintf("Date: %s -- %s\n", *startDate, *endDate))
	f.WriteString(fmt.Sprintf("Total: %d\n", *issuesResult.Total))
	for _, i := range issuesResult.Issues {
		note := fetchReleaseNoteFromPR(i)
		f.WriteString(note)
	}
	if *issuesResult.IncompleteResults {
		f.WriteString("!!Warning: Some release notes missing due to incomplete search result from github.")
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

func createQueryString(repo string) []string {
	var queries []string

	queries = addQuery(queries, "repo", *org, "/", repo)
	queries = addQuery(queries, "label", *label)
	queries = addQuery(queries, "is", "merged")
	queries = addQuery(queries, "type", "pr")
	queries = addQuery(queries, "merged", ">", *startDate)
	queries = addQuery(queries, "merged", "<", *endDate)

	return queries
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

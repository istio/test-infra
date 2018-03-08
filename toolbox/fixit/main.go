// Copyright 2018 Istio Authors
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
	"strings"
	"time"

	u "istio.io/test-infra/toolbox/util"
)

type orderV string
type sortV string

const (
	// Allowed field values can be found at https://developer.github.com/v3/search/#search-repositories
	desc    = orderV("desc")
	created = sortV("created")
)

var (
	org             = flag.String("user", "istio", "Github owner or org")
	tokenFile       = flag.String("token_file", "/Users/hklai/github_token", "Github token file (optional)")
	repo            = flag.String("repo", "istio", "Github repo")
	label            = flag.String("label", "kind/fixit", "Label to search for")
	sort            = flag.String("sort", string(created), "The sort field. Can be comments, created, or updated.")
	order           = flag.String("order", string(desc), "The sort order if sort parameter is provided. One of asc or desc.")
	startDate       = time.Date(2018, time.March, 5, 0, 0, 0, 0, time.Local)
	endDate       = time.Date(2018, time.March, 10, 0, 0, 0, 0, time.Local)
	gh              *u.GithubClient
)

func init() {
	flag.Parse()
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

// Increment a key in the given map.
func incrementCount(m map[string]int, key string) {
	val, ok := m[key]
	if !ok {
		val = 0
	}
	m[key] = val + 1
}

// Check if an event happened during the FixIt week.
func isFixItWeek(t time.Time) bool {
	return t.After(startDate) && t.Before(endDate)
}

// Find all metric related to issues
func findIssueMetric(metric *FixItMetric) {
	issueQueries := createIssueQuery("issue")
	log.Printf("Issue Query: %v", issueQueries)

	allIssues, err := gh.SearchIssues(issueQueries, *sort, *order)
	if err != nil {
		log.Printf("Failed to fetch Issues for %s: %s", repo, err)
		return
	}
	(*metric).totalIssues = len(allIssues)
	for _, issue := range allIssues {
		if issue.GetState() == "closed" {
			(*metric).totalClosedIssues++
		}
		events, err := gh.GetIssueEvents(*repo, (*issue).GetNumber())
		if err != nil {
			log.Printf("Failed to fetch events for issue %s: %s", (*issue).GetURL(), err)
			return
		}

		// Find the person who labeled the issue.
		for _, event := range events {
			if (*event).GetLabel() != nil {
				if (*event).GetLabel().GetName() == *label {
					incrementCount((*metric).issueLabeledMap, (*event).GetActor().GetLogin())
					break;
				}
			}
		}
		// Find the person who labeled the issue.
		for _, event := range events {
			if (*event).GetEvent() == "closed" && isFixItWeek(event.GetCreatedAt()) {
				login := (*event).GetActor().GetLogin()
				if login != "istio-merge-robot" {
					// Not counting the bot
					incrementCount((*metric).issueClosedMap, login)
					break;
				}
			}
		}
	}
}

// Find all metric related to pulls
func findPullMetric(metric *FixItMetric) {
	pullQueries := createIssueQuery("pr")
	log.Printf("Pull Query: %v", pullQueries)

	allPulls, err := gh.SearchIssues(pullQueries, *sort, *order)
	if err != nil {
		log.Printf("Failed to fetch PR for %s: %s", repo, err)
		return
	}
	for _, pull := range allPulls {
		if pull.GetState() == "closed" {
			incrementCount((*metric).pullClosedMap, (*pull).GetUser().GetLogin())
		} else if pull.GetState() == "open" {
			incrementCount((*metric).pullOpenMap, (*pull).GetUser().GetLogin())
		}
		reviews, err := gh.GetPullReviews(*repo, (*pull).GetNumber())
		if err != nil {
			log.Printf("Failed to fetch reviews for %s: %s", (*pull).GetURL(), err)
			return
		}

		for _, review := range reviews {
			// Multiple reviews in the same PR is counted as once
			reviewLogins := make(map[string]bool)
			if isFixItWeek((*review).GetSubmittedAt()) {
				reviewLogins[(*review).GetUser().GetLogin()] = true
			}
			for login, _ := range reviewLogins {
				incrementCount((*metric).pullReviewMap, login)
			}
		}
	}
}

// All metric that we capture.
type FixItMetric struct {
	totalIssues int
	totalClosedIssues int
	issueLabeledMap map[string]int
	issueClosedMap map[string]int
	pullClosedMap map[string]int
	pullOpenMap map[string]int
	pullReviewMap map[string]int
}

// Building a new FixItMetric with all maps initialized.
func NewFixItMetric() *FixItMetric {
	var metric FixItMetric
	metric.issueLabeledMap = make(map[string]int)
	metric.issueClosedMap = make(map[string]int)
	metric.pullClosedMap = make(map[string]int)
	metric.pullOpenMap = make(map[string]int)
	metric.pullReviewMap = make(map[string]int)
	return &metric
}

func main() {
	metric := NewFixItMetric()
	findIssueMetric(metric)
	findPullMetric(metric)
	printReport(metric)
}

func printReport(metric *FixItMetric) {
	fmt.Println("==================== Istio FixIt 2018 ====================")
	fmt.Printf("Total number of issues: %d", metric.totalIssues)
	fmt.Println()
	fmt.Printf("Total number of closed issues: %d\n", metric.totalClosedIssues)
	fmt.Println()

	printLeaderBoard("Total Issues Labeled", metric.issueLabeledMap)
	printLeaderBoard("Total Issues Closed", metric.issueClosedMap)
	printLeaderBoard("Total Pulls Closed", metric.pullClosedMap)
	printLeaderBoard("Total Open Pulls", metric.pullOpenMap)
	printLeaderBoard("Total Pull Reviews", metric.pullReviewMap)

	fmt.Println("==========================================================")
}

func printLeaderBoard(title string, m map[string]int) {
	fmt.Println(title)
	fmt.Println("======================")

	pairList := u.SortMapByValue(m)
	for i := len(pairList) - 1; i >= 0; i-- {
		kv := pairList[i]
		fmt.Printf("%s\t%d\n", kv.Key, kv.Value)
	}
	fmt.Println()
}

func createIssueQuery(issuetype string) []string {
	var queries []string
	queries = addQuery(queries, "repo", *org, "/", *repo)
	queries = addQuery(queries, "label", *label)
	queries = addQuery(queries, "type", issuetype)
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


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
	"os"
	"regexp"
	"strings"

	u "istio.io/test-infra/toolbox/util"
)

var (
	org       = flag.String("user", "istio", "Github owner or org")
	tokenFile = flag.String("token_file", "", "Github token file. May run into rate limiting if not provided.")
	repos     = flag.String("repos", "istio,api,proxy", "Github repos, separate using \",\"")
	sort      = flag.String("sort", "merged", "The sort field. Can be comments, merged, created, updated, etc.")
	order     = flag.String("order", "desc", "The sort order. One of asc or desc.")
	startDate = flag.String("start_date", "2018-08-13", "The start date when the PR was merged")
	branch    = flag.String("branch", "release-1.0", "Branch name")
	outDir    = flag.String("out_dir", "/tmp", "Output file path")
	gh        *u.GithubClient
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

func findPulls(repo string) {
	filename := fmt.Sprintf("%s/%s.csv", *outDir, repo)
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal("Cannot create file", err)
		return
	}
	log.Printf("Created %s", filename)

	pullQueries := createIssueQuery(repo)
	log.Printf("Pull Query: %v", pullQueries)

	allPulls, err := gh.SearchIssues(pullQueries, *sort, *order)
	if err != nil {
		log.Printf("Failed to fetch PR for %s: %s", repo, err)
		return
	}

	// Insert CSV header
	_, _ = fmt.Fprintf(file, "PR (Ordered by descending merge time), Description, Author, PR in master\n")

	for _, pull := range allPulls {
		title := pull.GetTitle()
		masterPr := ""
		// Find corresponding PR in master if this is a cherrypick.
		re := regexp.MustCompile(`\(#(.*?)\)`)
		match := re.FindStringSubmatch(title)
		if match != nil {
			masterPr = fmt.Sprintf("https://github.com/istio/%s/pull/%s", repo, match[1])
		}

		// Master uses its own set of SHA (pointing to master of upstream repo)
		titleLower := strings.ToLower(title)
		if strings.Contains(titleLower, "update envoy sha") ||
			strings.Contains(titleLower, "update proxy sha") ||
			strings.Contains(titleLower, "update api sha") {
			masterPr = "NOT NEEDED"
		}

		_, _ = fmt.Fprintf(file, "%s, \"%s\", %s, %s\n", pull.GetHTMLURL(), title, pull.GetUser().GetLogin(), masterPr)
	}
	err = file.Close()
	if err != nil {
		log.Printf("Failed to close file %s", err)
		return
	}
}

func main() {
	for _, repo := range strings.Split(*repos, ",") {
		findPulls(repo)
	}
}

func createIssueQuery(repo string) []string {
	var queries []string
	queries = append(queries, fmt.Sprintf("repo:%s/%s", *org, repo))
	queries = append(queries, "type:pr")
	queries = append(queries, "merged:>="+*startDate)
	queries = append(queries, "base:"+*branch)

	return queries
}

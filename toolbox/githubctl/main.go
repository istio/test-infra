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
	"log"

	"istio.io/test-infra/toolbox/util"
)

var (
	owner      = flag.String("owner", "istio", "Github owner or org")
	tokenFile  = flag.String("token_file", "", "File containing Github API Access Token")
	op         = flag.String("op", "", "Operation to be performed")
	repo       = flag.String("repo", "", "Repository to which op is applied")
	baseBranch = flag.String("base_branch", "", "Branch to which op is applied")
	refSHA     = flag.String("ref_sha", "", "Reference commit SHA used to update base branch")
	githubClnt *util.GithubClient
)

// Panic if value not specified
func assertNotEmpty(name string, value *string) {
	if value == nil || *value == "" {
		log.Panicf("%s must be specified\n", name)
	}
}

func fastForward(repo, baseBranch, refSHA *string) error {
	assertNotEmpty("repo", repo)
	assertNotEmpty("base_branch", baseBranch)
	assertNotEmpty("ref_sha", refSHA)
	return githubClnt.FastForward(*repo, *baseBranch, *refSHA)
}

func init() {
	flag.Parse()
	assertNotEmpty("token_file", tokenFile)
	token, err := util.GetAPITokenFromFile(*tokenFile)
	if err != nil {
		log.Panicf("Error accessing user supplied token_file: %v\n", err)
	}
	githubClnt = util.NewGithubClient(*owner, token)
}

func main() {
	switch *op {
	case "fastForward":
		if err := fastForward(repo, baseBranch, refSHA); err != nil {
			log.Printf("Error during fastForward: %v\n", err)
		}
	default:
		log.Printf("Unsupported operation: %s\n", *op)
	}
}

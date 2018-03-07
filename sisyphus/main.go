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

	s "istio.io/test-infra/sisyphus"
)

const (
	// Message Info
	sender         = "istio.testing@gmail.com"
	oncallMaillist = "istio-oncall@googlegroups.com"
	subject        = "ATTENTION - Istio Post-Submit Test Failed"
	prologue       = "Hi istio-oncall,\n\n" +
		"Post-Submit is failing in istio/istio, please take a look at following failure(s) and fix ASAP\n\n"
	epilogue = "\nIf you have any questions about this message or notice inaccuracy, please contact istio-engprod@google.com."

	// Prow result GCS
	lastBuildTXT  = "latest-build.txt"
	finishedJSON  = "finished.json"
	startedJSON   = "started.json"
	gubernatorURL = "https://k8s-gubernator.appspot.com/build/istio-prow"

	// Token and password file
	tokenFileDocker        = "/etc/github/git-token"
	gmailAppPassFileDocker = "/etc/gmail/gmail-app-pass"
	identity               = "istio-bot"

	// Prow GCP settings
	prowProject = "istio-testing"
	prowZone    = "us-west1-a"
)

var (
	gcsBucket            = flag.String("bucket", "istio-prow", "Prow artifact GCS bucket name.")
	interval             = flag.Int("interval", 60, "Check and report interval(seconds)")
	numRerun             = flag.Int("num_rerun", 3, "Number of reruns to detect flakyness")
	owner                = flag.String("owner", "istio", "Github owner or org")
	tokenFile            = flag.String("github_token", tokenFileDocker, "Path to github token")
	gmailAppPassFile     = flag.String("gmail_app_password", gmailAppPassFileDocker, "Path to gmail application password")
	protectedRepo        = flag.String("protected_repo", "istio", "Protected repo")
	protectedBranch      = flag.String("protected_branch", "master", "Protected branch")
	guardProtectedBranch = flag.Bool("guard", false, "Suspend merge bot if postsubmit fails")
	emailSending         = flag.Bool("email_sending", true, "Sending alert email")
	catchFlakesByRun     = flag.Bool("catch_flakes_by_rerun", true, "whether to rerun failed jobs to detect flakyness")

	protectedJobs = []string{"istio-postsubmit", "e2e-suite-rbac-auth", "e2e-suite-rbac-no_auth"}
)

func main() {
	flag.Parse()
	sisyphusd := s.SisyphusDaemon(protectedJobs, prowProject, prowZone)
	if *emailSending {
		gmailAppPass, err := u.GetPasswordFromFile(*gmailAppPassFile)
		if err != nil {
			log.Fatalf("Error accessing gmail app password: %v", err)
		}
		sisyphusd.SetAlert(gmailAppPass, identity, sender, oncallMaillist, &s.AlertConfig{
			Subject:  subject,
			Prologue: prologue,
			Epilogue: epilogue,
		})
	}
	sisyphusd.Start()
}

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
)

const (
	// Message Info
	sender          = "istio.testing@gmail.com"
	oncallMaillist  = "istio-oncall@googlegroups.com"
	messageSubject  = "ATTENTION - Istio Post-Submit Test Failed"
	messagePrologue = "Hi istio-oncall,\n\n" +
		"Post-Submit is failing in istio/istio, please take a look at following failure(s) and fix ASAP\n\n"
	messageEnding = "\nIf you have any questions about this message or notice inaccuracy, please contact istio-engprod@google.com."
	losAngeles    = "America/Los_Angeles"

	// Gmail setting
	gmailSMTPSERVER = "smtp.gmail.com"
	gmailSMTPPORT   = 587

	// Prow result GCS
	lastBuildTXT  = "latest-build.txt"
	flakeStatJSON = "flakeStat.json"
	finishedJSON  = "finished.json"
	startedJSON   = "started.json"
	gubernatorURL = "https://k8s-gubernator.appspot.com/build/istio-prow"

	// Token and password file
	tokenFileDocker        = "/etc/github/git-token"
	gmailAppPassFileDocker = "/etc/gmail/gmail-app-pass"
)

var (
	gcsBucket        = flag.String("bucket", "istio-prow", "Prow artifact GCS bucket name.")
	interval         = flag.Int("interval", 60, "Check and report interval(seconds)")
	numRerun         = flag.Int("num_rerun", 1, "Number of reruns to detect flakyness")
	owner            = flag.String("owner", "istio", "Github owner or org")
	tokenFile        = flag.String("github_token", tokenFileDocker, "Path to github token")
	gmailAppPassFile = flag.String("gmail__app_password", gmailAppPassFileDocker, "Path to gmail application password")
	protectedRepo    = flag.String("protected_repo", "istio", "Protected repo")
	protectedBranch  = flag.String("protected_branch", "master", "Protected branch")
	guardMaster      = flag.Bool("guard_master", false, "Suspend merge bot if postsubmit fails")
	emailSending     = flag.Bool("email_sending", false, "Sending alert email")
	catchFlakesByRun = flag.Bool("catch_flakes_by_rerun", true, "whether to rerun failed jobs to detect flakyness")
)

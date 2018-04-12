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
	"context"
	"flag"
	"log"

	"istio.io/test-infra/sisyphus"
	u "istio.io/test-infra/toolbox/util"
)

const (
	// Alert settings
	sender         = "istio.testing@gmail.com"
	oncallMaillist = "istio-oncall@googlegroupsisyphus.com"
	subject        = "ATTENTION - Istio Post-Submit Test Failed"
	prologue       = "Hi istio-oncall,\n\n" +
		"Post-Submit is failing in istio/istio, please take a look at following failure(s) and fix ASAP\n\n"
	epilogue = "\nIf you have any questions about this message or notice inaccuracy, please contact istio-engprod@google.com."
	identity = "istio-bot"

	// Prow
	prowProject   = "istio-testing"
	prowZone      = "us-west1-a"
	gubernatorURL = "https://k8s-gubernator.appspot.com/build/istio-prow"
	gcsBucket     = "istio-prow"

	// Branch protection
	owner           = "istio"
	protectedRepo   = "istio"
	protectedBranch = "master"
)

var (
	tokenFile            = flag.String("github_token", "/etc/github/git-token", "Path to github token")
	gmailAppPassFile     = flag.String("gmail_app_password", "/etc/gmail/gmail-app-pass", "Path to gmail application password")
	guardProtectedBranch = flag.Bool("guard", false, "Suspend merge bot if postsubmit fails")
	emailSending         = flag.Bool("email_sending", false, "Sending alert email")
	catchFlakesByRun     = flag.Bool("catch_flakes_by_rerun", true, "whether to rerun failed jobs to detect flakyness")

	protectedJobs = []string{
		"daily-e2e-cluster_wide-auth-default",
		"daily-e2e-cluster_wide-auth-skew",
		"daily-e2e-cluster_wide-auth",
		"daily-e2e-rbac-auth-default",
		"daily-e2e-rbac-auth-skew",
		"daily-e2e-rbac-auth",
		"daily-e2e-rbac-no_auth-default",
		"daily-e2e-rbac-no_auth-skew",
		"daily-e2e-rbac-no_auth",
	}
	presubmitJobs = protectedJobs
)

func init() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// Connect to the Prow cluster
	if _, err := u.Shell(`gcloud container clusters get-credentials prow \
		--project=%s --zone=%s`, prowProject, prowZone); err != nil {
		log.Fatalf("Unable to switch to prow cluster: %v\n", err)
	}
}

func main() {
	gcsClient := u.NewGCSClient(gcsBucket)
	sisyphusd := sisyphus.NewDaemonUsingProw(
		protectedJobs, presubmitJobs, prowProject, prowZone, gubernatorURL,
		gcsBucket,
		gcsClient,
		sisyphus.NewStorage(),
		&sisyphus.Config{
			CatchFlakesByRun: *catchFlakesByRun,
		})
	if *emailSending {
		gmailAppPass, err := u.GetPasswordFromFile(*gmailAppPassFile)
		if err != nil {
			log.Fatalf("Error accessing gmail app password: %v", err)
		}
		if err := sisyphusd.SetAlert(gmailAppPass, identity, sender, oncallMaillist,
			&sisyphus.AlertConfig{
				Subject:  subject,
				Prologue: prologue,
				Epilogue: epilogue,
			}); err != nil {
			log.Fatalf("Failed to set up alerts: %v", err)
		}
	}
	if *guardProtectedBranch {
		token, err := u.GetAPITokenFromFile(*tokenFile)
		if err != nil {
			log.Fatalf("Error accessing user supplied token_file: %v\n", err)
		}
		sisyphusd.SetProtectedBranch(owner, token, protectedRepo, protectedBranch)
	}
	sisyphusd.Start(context.Background())
}

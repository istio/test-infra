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

package util

import (
	"github.com/google/go-github/github"
)

var (
	ci = NewCIState()
)

// CIState defines constants representing possible states of
// continuous integration tests
type CIState struct {
	Success string
	Failure string
	Pending string
	Error   string
}

// NewCIState creates a new CIState
func NewCIState() *CIState {
	return &CIState{
		Success: "success",
		Failure: "failure",
		Pending: "pending",
		Error:   "error",
	}
}

// GetCIState does NOT trust the given combined output but instead walk
// through the CI results, count states, and determine the final state
// as either pending, failure, or success
func GetCIState(combinedStatus *github.CombinedStatus, skipContext func(string) bool) string {
	var failures, pending, successes int
	for _, status := range combinedStatus.Statuses {
		if *status.State == ci.Error || *status.State == ci.Failure {
			if skipContext != nil && skipContext(*status.Context) {
				continue
			}
			failures++
		} else if *status.State == ci.Pending {
			pending++
		} else if *status.State == ci.Success {
			successes++
		} else {
			log.Printf("Check Status %s is unknown", *status.State)
		}
	}
	if pending > 0 {
		return ci.Pending
	} else if failures > 0 {
		return ci.Failure
	} else {
		return ci.Success
	}
}

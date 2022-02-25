// Copyright Istio Authors
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

package decorator

import (
	"log"

	prowjob "k8s.io/test-infra/prow/apis/prowjobs/v1"
	"k8s.io/test-infra/prow/config"
)

const (
	ModifierHidden            = "hidden"
	ModifierPresubmitOptional = "presubmit_optional"
	ModifierPresubmitSkipped  = "presubmit_skipped"
)

func ApplyModifiersPresubmit(presubmit *config.Presubmit, jobModifiers []string) {
	for _, modifier := range jobModifiers {
		switch modifier {
		case ModifierPresubmitOptional:
			presubmit.Optional = true
		case ModifierHidden:
			presubmit.SkipReport = true
			presubmit.ReporterConfig = &prowjob.ReporterConfig{
				Slack: &prowjob.SlackReporterConfig{
					JobStatesToReport: []prowjob.ProwJobState{},
				},
			}
		case ModifierPresubmitSkipped:
			presubmit.AlwaysRun = false
		default:
			log.Fatalf("Modifier %q is not unsupported", modifier)
		}
	}
}

func ApplyModifiersPostsubmit(postsubmit *config.Postsubmit, jobModifiers []string) {
	for _, modifier := range jobModifiers {
		switch modifier {
		case ModifierPresubmitOptional, ModifierPresubmitSkipped:
			// No effect on postsubmit
		case ModifierHidden:
			postsubmit.SkipReport = true
			f := false
			postsubmit.ReporterConfig = &prowjob.ReporterConfig{
				Slack: &prowjob.SlackReporterConfig{
					Report: &f,
				},
			}
		default:
			log.Fatalf("Modifier %q is not unsupported", modifier)
		}
	}
}

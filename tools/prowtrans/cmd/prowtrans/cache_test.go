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

package main

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/prow/pkg/config"

	"istio.io/test-infra/tools/prowtrans/pkg/configuration"
)

func TestUpdateInitContainerEnvs(t *testing.T) {
	job := &config.JobBase{
		Spec: &v1.PodSpec{
			Containers: []v1.Container{{
				Name: "test",
			}},
			InitContainers: []v1.Container{{
				Name: "bazel-remote",
				Env:  []v1.EnvVar{{Name: "BAZEL_REMOTE_BUCKET", Value: "public-cache"}},
			}, {
				Name: "setup",
				Env:  []v1.EnvVar{{Name: "BAZEL_REMOTE_BUCKET", Value: "unchanged"}},
			}},
		},
	}
	privateBucket := "private-cache"

	updateInitContainerEnvs(options{Transform: configuration.Transform{
		InitContainerEnv: map[string]map[string]*string{
			"bazel-remote": {"BAZEL_REMOTE_BUCKET": &privateBucket},
		},
	}}, job)

	if got := job.Spec.Containers[0].Env; len(got) != 0 {
		t.Fatalf("primary container environment = %v, want unchanged", got)
	}
	if got := job.Spec.InitContainers[0].Env; len(got) != 1 || got[0].Value != privateBucket {
		t.Fatalf("sidecar environment = %v, want private bucket", got)
	}
	if got := job.Spec.InitContainers[1].Env; len(got) != 1 || got[0].Value != "unchanged" {
		t.Fatalf("unselected init container environment = %v, want unchanged", got)
	}
}

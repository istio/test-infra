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
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/prow/pkg/config"

	"istio.io/test-infra/tools/prowgen/pkg/spec"
)

func TestRequirementOverridesNodeSelector(t *testing.T) {
	podSpec := &v1.PodSpec{
		NodeSelector: map[string]string{
			"kubernetes.io/arch": "amd64",
			"testing":            "test-pool",
		},
		Tolerations: []v1.Toleration{
			{
				Key:      "kubernetes.io/arch",
				Operator: v1.TolerationOpEqual,
				Value:    "arm64",
				Effect:   v1.TaintEffectNoSchedule,
			},
		},
	}

	requirement := spec.RequirementPreset{
		PodSpec: &v1.PodSpec{
			NodeSelector: map[string]string{
				"testing": "trusted",
			},
			Tolerations: []v1.Toleration{
				{
					Key:      "testing",
					Operator: v1.TolerationOpEqual,
					Value:    "trusted",
					Effect:   v1.TaintEffectNoSchedule,
				},
			},
		},
	}

	mergeRequirement(nil, nil, podSpec, nil, nil, requirement)

	want := map[string]string{
		"kubernetes.io/arch": "amd64",
		"testing":            "trusted",
	}
	if diff := cmp.Diff(want, podSpec.NodeSelector); diff != "" {
		t.Fatalf("unexpected node selector (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]v1.Toleration{
		{
			Key:      "kubernetes.io/arch",
			Operator: v1.TolerationOpEqual,
			Value:    "arm64",
			Effect:   v1.TaintEffectNoSchedule,
		},
		{
			Key:      "testing",
			Operator: v1.TolerationOpEqual,
			Value:    "trusted",
			Effect:   v1.TaintEffectNoSchedule,
		},
	}, podSpec.Tolerations); diff != "" {
		t.Fatalf("unexpected tolerations (-want +got):\n%s", diff)
	}
}

func TestRequirementAddsSidecarWithoutSharingSecrets(t *testing.T) {
	job := &config.JobBase{
		Spec: &v1.PodSpec{
			Containers: []v1.Container{{Name: "test"}},
		},
	}
	requirements := map[string]spec.RequirementPreset{
		"credentials": {
			Secrets: []spec.Secret{{Name: "cache-credentials"}},
		},
		"cache": {
			Env:      []v1.EnvVar{{Name: "CACHE_URL", Value: "http://localhost:8080"}},
			Sidecars: []v1.Container{{Name: "cache"}},
		},
	}

	ApplyRequirements(spec.BaseConfig{}, job, []string{"cache", "credentials"}, nil, requirements)

	if len(job.Spec.Containers) != 1 {
		t.Fatalf("got %d primary containers, want 1", len(job.Spec.Containers))
	}
	if got := job.Spec.Containers[0].Env; len(got) != 2 || got[0].Name != "CACHE_URL" || got[1].Name != "GCP_SECRETS" {
		t.Fatalf("primary container env = %v, want cache URL and secrets", got)
	}
	if len(job.Spec.InitContainers) != 1 {
		t.Fatalf("got %d sidecars, want 1", len(job.Spec.InitContainers))
	}
	if got := job.Spec.InitContainers[0].Env; len(got) != 0 {
		t.Fatalf("sidecar env = %v, want no inherited environment", got)
	}
	if got := job.Spec.InitContainers[0].RestartPolicy; got == nil || *got != v1.ContainerRestartPolicyAlways {
		t.Fatalf("sidecar restart policy = %v, want Always", got)
	}
}

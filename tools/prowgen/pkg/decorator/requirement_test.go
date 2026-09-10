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

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

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
)

func TestFilterDuplicateEnvVars(t *testing.T) {
	secretRef := &v1.EnvVarSource{
		SecretKeyRef: &v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{Name: "token"},
			Key:                  "value",
		},
	}

	tests := []struct {
		name string
		env  []v1.EnvVar
		want []v1.EnvVar
	}{
		{
			name: "empty",
			env:  nil,
			want: nil,
		},
		{
			name: "no duplicates keeps order",
			env: []v1.EnvVar{
				{Name: "B", Value: "1"},
				{Name: "A", Value: "2"},
				{Name: "C", Value: "3"},
			},
			want: []v1.EnvVar{
				{Name: "B", Value: "1"},
				{Name: "A", Value: "2"},
				{Name: "C", Value: "3"},
			},
		},
		{
			name: "duplicates are removed",
			env: []v1.EnvVar{
				{Name: "BUILD_WITH_CONTAINER", Value: "0"},
				{Name: "BUILD_WITH_CONTAINER", Value: "0"},
				{Name: "FOO", Value: "bar"},
			},
			want: []v1.EnvVar{
				{Name: "BUILD_WITH_CONTAINER", Value: "0"},
				{Name: "FOO", Value: "bar"},
			},
		},
		{
			name: "last definition wins at the first position",
			env: []v1.EnvVar{
				{Name: "FOO", Value: "base"},
				{Name: "BAR", Value: "bar"},
				{Name: "FOO", Value: "override"},
			},
			want: []v1.EnvVar{
				{Name: "FOO", Value: "override"},
				{Name: "BAR", Value: "bar"},
			},
		},
		{
			name: "valueFrom is preserved",
			env: []v1.EnvVar{
				{Name: "TOKEN", Value: "placeholder"},
				{Name: "TOKEN", ValueFrom: secretRef},
			},
			want: []v1.EnvVar{
				{Name: "TOKEN", ValueFrom: secretRef},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, filterDuplicateEnvVars(tt.env)); diff != "" {
				t.Errorf("filterDuplicateEnvVars() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

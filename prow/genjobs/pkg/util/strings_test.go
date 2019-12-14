/*
Copyright 2019 Istio Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"testing"
)

func TestNormalizeHost(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sep      string
		expected string
	}{
		{
			name:     "normalize 3-part org w/ host",
			input:    "https://go.google.com/go/json/encode",
			sep:      ".",
			expected: "go.json.encode",
		},
		{
			name:     "normalize 2-part org w/ host",
			input:    "https://sub.google.com/eng/code",
			sep:      ".",
			expected: "eng.code",
		},
		{
			name:     "normalize 1-part org w/ host",
			input:    "https://gerrit.googlesource.com/istio",
			sep:      ".",
			expected: "istio",
		},
		{
			name:     "normalize 0-part org w/ host",
			input:    "https://gerrit.googlesource.com",
			sep:      ".",
			expected: "",
		},
		{
			name:     "normalize http host",
			input:    "http://gerrit.googlesource.com",
			sep:      ".",
			expected: "",
		},
		{
			name:     "normalize simple host",
			input:    "https://abc",
			sep:      ".",
			expected: "",
		},
		{
			name:     "normalize simple scheme",
			input:    "https://",
			sep:      ".",
			expected: "",
		},
		{
			name:     "normalize org only",
			input:    "istio",
			sep:      ".",
			expected: "istio",
		},
		{
			name:     "trim spaces",
			input:    "https://gerrit.googlesource.com/istio    ",
			sep:      ".",
			expected: "istio",
		},
		{
			name:     "trim backslashes",
			input:    "https://gerrit.googlesource.com/istio/",
			sep:      ".",
			expected: "istio",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := NormalizeOrg(test.input, test.sep)

			if actual != test.expected {
				t.Errorf("Actual: %v ; Expected: %v", actual, test.expected)
			}
		})
	}
}

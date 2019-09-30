// Copyright 2019 Istio Authors
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

package config

import (
	"fmt"
	"os"
	"testing"
)

func TestGenerateConfig(t *testing.T) {
	tests := []string{"simple"}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			jobs := ReadJobConfig(fmt.Sprintf("testdata/%s.yaml", tt))
			for _, branch := range jobs.Branches {
				output := ConvertJobConfig(jobs, branch)
				if os.Getenv("REFRESH_GOLDEN") == "true" {
					WriteConfig(output, fmt.Sprintf("testdata/%s.gen.yaml", tt))
				}
				if err := CheckConfig(output, fmt.Sprintf("testdata/%s.gen.yaml", tt)); err != nil {
					t.Fatal(err.Error())
				}
			}
		})
	}
}

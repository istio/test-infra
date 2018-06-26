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

package config

import (
	"testing"

	"k8s.io/test-infra/prow/config"
	_ "k8s.io/test-infra/prow/hook"
	"k8s.io/test-infra/prow/plugins"
)

func TestConfig(t *testing.T) {
	if _, err := config.Load("../config.yaml", "../cluster/jobs/"); err != nil {
		t.Fatalf("could not read configs: %v", err)
	}
	// TODO: Add branch protection validation once its in
}

// Make sure that our plugins are valid.
func TestPlugins(t *testing.T) {
	pa := &plugins.PluginAgent{}
	if err := pa.Load("../plugins.yaml"); err != nil {
		t.Fatalf("could not load plugins: %v", err)
	}
}

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

package testgrids

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/GoogleCloudPlatform/testgrid/config"
	config_pb "github.com/GoogleCloudPlatform/testgrid/pb/config"
	prow_config "k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/testgrid/pkg/configurator/configurator"
	"k8s.io/test-infra/testgrid/pkg/configurator/options"

	configflagutil "k8s.io/test-infra/prow/flagutil/config"
)

var (
	dashboardPrefixes = []string{
		"istio",
	}
)

var defaultInputs options.MultiString = []string{"."}
var prowPath = flag.String("prow-config", "../prow/config.yaml", "Path to prow config")
var jobPath = flag.String("job-config", "../prow/cluster/jobs", "Path to prow job config")
var defaultYAML = flag.String("default", "./default.yaml", "Default yaml for testgrid")
var inputs options.MultiString
var protoPath = flag.String("config", "", "Path to TestGrid config proto")

// Shared testgrid config, loaded at TestMain.
var cfg *config_pb.Configuration

// Shared prow config, loaded at Test Main
var prowConfig *prow_config.Config

func TestMain(m *testing.M) {
	flag.Var(&inputs, "yaml", "comma-separated list of input YAML files or directories")
	flag.Parse()
	if *protoPath == "" {
		if len(inputs) == 0 {
			inputs = defaultInputs
		}
		// Generate proto from testgrid config
		tmpDir, err := os.MkdirTemp("", "testgrid-config-test")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer os.RemoveAll(tmpDir)
		tmpFile := path.Join(tmpDir, "test-proto")

		opt := options.Options{
			Inputs: inputs,
			ProwConfig: configflagutil.ConfigOptions{
				ConfigPath:    *prowPath,
				JobConfigPath: *jobPath,
			},
			DefaultYAML:     *defaultYAML,
			Output:          flagutil.NewStringsBeenSet(tmpFile),
			Oneshot:         true,
			StrictUnmarshal: true,
		}

		if err := configurator.RealMain(&opt); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		protoPath = &tmpFile
	}

	var err error
	cfg, err = config.Read(context.Background(), *protoPath, nil)
	if err != nil {
		fmt.Printf("Could not load config: %v\n", err)
		os.Exit(1)
	}

	prowConfig, err = prow_config.Load(*prowPath, *jobPath, nil, "")
	if err != nil {
		fmt.Printf("Could not load prow configs: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestConfig(t *testing.T) {
	dashboardNames := sets.NewString()
	dashboardGroupNames := sets.NewString()
	dashboardToGroupMap := make(map[string]string)

	for _, db := range cfg.Dashboards {
		dashboardNames.Insert(db.Name)

	}
	for _, dbg := range cfg.DashboardGroups {
		dashboardGroupNames.Insert(dbg.Name)
		for _, dashboard := range dbg.DashboardNames {
			dashboardToGroupMap[dashboard] = dbg.Name
		}
	}

	// Convention: all dashboard (group) names must start with a well known prefix
	names := sets.NewString()
	names = names.Union(dashboardNames)
	names = names.Union(dashboardGroupNames)
	for name := range names {
		found := false
		for _, prefix := range dashboardPrefixes {
			if strings.HasPrefix(name, prefix) || name == prefix {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Dashboard/DashboardGroup %v: must prefix with one of: %v", name, dashboardPrefixes)
		}

	}

	// Convention: all dashboards must be under a dashboard group
	// Convention: dashboards should live under the dashboard group with the longest common prefix
	// (e.g. project-foo-bar dashboard should be in project-foo group over project group).
	for dashboard := range dashboardNames {
		assignedGroup, ok := dashboardToGroupMap[dashboard]
		if !ok {
			t.Errorf("Dashboard %v: must be in a dashboard_group", dashboard)
		}
		longestGroupPrefix := ""
		for thisGroup := range dashboardGroupNames {
			if strings.HasPrefix(dashboard, thisGroup) && len(thisGroup) > len(longestGroupPrefix) {
				longestGroupPrefix = thisGroup
			}
		}
		if longestGroupPrefix == "" {
			t.Errorf("Dashboard %v: should be in a dashboard_group with a common prefix instead of dashboard_group '%v'", dashboard, assignedGroup)
		} else if assignedGroup != longestGroupPrefix {
			t.Errorf("Dashboard %v: should be in dashboard_group '%v' instead of dashboard_group '%v'", dashboard, longestGroupPrefix, assignedGroup)
		}
	}
}

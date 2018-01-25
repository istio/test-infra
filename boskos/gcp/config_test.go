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

package gcp

import (
	"reflect"
	"testing"

	"k8s.io/test-infra/boskos/mason"
	"k8s.io/test-infra/boskos/ranch"
)

func TestParseInvalidConfig(t *testing.T) {
	expected := resourcesConfig{
		ProjectConfigs: []projectConfig{
			{
				Type: "type1",
				Clusters: []clusterConfig{
					{
						MachineType: "n1-standard-2",
						NumNodes:    4,
						Version:     "1.7",
						Zone:        "us-central-1f",
					},
				},
				Vms: []virtualMachineConfig{
					{
						MachineType: "n1-standard-4",
						SourceImage: "projects/debian-cloud/global/images/debian-9-stretch-v20180105",
						Zone:        "us-central-1f",
						Tags: []string{
							"http-server",
							"https-server",
						},
						Scopes: []string{
							"https://www.googleapis.com/auth/cloud-platform",
						},
					},
				},
			},
		},
	}
	conf, err := mason.ParseConfig("test-configs.yaml")
	if err != nil {
		t.Error("could not parse config")
	}
	config, err := ConfigConverter(conf[0].Config.Content)
	if err != nil {
		t.Errorf("cannot parse object")
	} else {
		if !reflect.DeepEqual(expected, *config.(*resourcesConfig)) {
			t.Error("Object differ")
		}
	}
}

func TestParseConfig(t *testing.T) {
	configs, err := mason.ParseConfig("../configs.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	resources, err := ranch.ParseConfig("../resources.yaml")
	if err != nil {
		t.Errorf(err.Error())
	}
	if err = mason.ValidateConfig(configs, resources); err != nil {
		t.Errorf(err.Error())
	}
}

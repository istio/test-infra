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

package gcp

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"k8s.io/test-infra/boskos/common"
	"k8s.io/test-infra/boskos/mason"
	"k8s.io/test-infra/boskos/ranch"
)

func TestParseInvalidConfig(t *testing.T) {
	expected := resourcesConfig{
		"type1": {{
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
		}},
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
	} else {

		for _, config := range configs {
			switch config.Config.Type {
			case ResourceConfigType:
				m, err := ConfigConverter(config.Config.Content)
				if err != nil {
					t.Errorf("unable to parse config %s %v", config.Name, err)
				}
				needs := common.ResourceNeeds{}
				rc, ok := m.(*resourcesConfig)
				if !ok {
					t.Errorf("cannot convert masonable to resourceConfig")
				} else {
					for rType, reqs := range *rc {
						needs[rType] = len(reqs)
					}
					if !reflect.DeepEqual(needs, config.Needs) {
						t.Errorf("Needs do not match for config %s. Expected %v found %v", config.Name, config.Needs, needs)
					}
				}

			}

		}
	}

	resources, err := ranch.ParseConfig("../resources.yaml")
	if err != nil {
		t.Errorf(err.Error())
	}
	if err = mason.ValidateConfig(configs, resources); err != nil {
		t.Errorf(err.Error())
	}
}

type faker struct {
	fail     bool
	waitTime time.Duration
}

func (f faker) do(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(f.waitTime):
	}
	if f.fail {
		return fmt.Errorf("fail")
	}
	return nil
}

type fakeVMCreator struct {
	f *faker
}

func (vmc *fakeVMCreator) create(ctx context.Context, p string, c virtualMachineConfig) (*instanceInfo, error) {
	if vmc.f != nil {
		if err := vmc.f.do(ctx); err != nil {
			return nil, err
		}
	}
	return &instanceInfo{
		Name: "VMname",
		Zone: "VMzone",
	}, nil
}

type fakeClusterCreator struct {
	f *faker
}

func (cc *fakeClusterCreator) create(ctx context.Context, p string, c clusterConfig) (*instanceInfo, error) {
	if cc.f != nil {
		if err := cc.f.do(ctx); err != nil {
			return nil, err
		}
	}
	return &instanceInfo{
		Name: "ClusterName",
		Zone: "ClusterZone",
	}, nil
}

func TestResourcesConfig_Construct(t *testing.T) {
	type expected struct {
		err  string
		info *ResourceInfo
	}

	testCases := []struct {
		name    string
		rc      resourcesConfig
		res     *common.Resource
		types   common.TypeToResources
		result  expected
		vmf, cf *faker
	}{
		{
			name: "success",
			rc: resourcesConfig{
				"test": {{
					Clusters: []clusterConfig{
						{},
					},
					Vms: []virtualMachineConfig{
						{},
					},
				}},
			},
			res: &common.Resource{
				Name: "test",
			},
			types: common.TypeToResources{
				"test": []*common.Resource{
					{Name: "leased"},
				},
			},
			result: expected{
				info: &ResourceInfo{
					"leased": {
						Clusters: []instanceInfo{
							{
								Name: "ClusterName",
								Zone: "ClusterZone",
							},
						},
						VMs: []instanceInfo{
							{
								Name: "VMname",
								Zone: "VMzone",
							},
						},
					},
				},
			},
		},
		{
			name: "timeout vm creation",
			rc: resourcesConfig{
				"test": {{
					Clusters: []clusterConfig{
						{},
					},
					Vms: []virtualMachineConfig{
						{},
					}},
				},
			},
			res: &common.Resource{
				Name: "test",
			},
			types: common.TypeToResources{
				"test": []*common.Resource{
					{Name: "leased"},
				},
			},
			result: expected{
				err: "context deadline exceeded",
			},
			vmf: &faker{
				waitTime: 2 * time.Second,
			},
		},
		{
			name: "timeout cluster creation",
			rc: resourcesConfig{
				"test": {{
					Clusters: []clusterConfig{
						{},
					},
					Vms: []virtualMachineConfig{
						{},
					}},
				},
			},
			res: &common.Resource{
				Name: "test",
			},
			types: common.TypeToResources{
				"test": []*common.Resource{
					{Name: "leased"},
				},
			},
			result: expected{
				err: "context deadline exceeded",
			},
			cf: &faker{
				waitTime: 2 * time.Second,
			},
		},
		{
			name: "failed vm creation",
			rc: resourcesConfig{
				"test": {{
					Clusters: []clusterConfig{
						{},
					},
					Vms: []virtualMachineConfig{
						{},
					},
				}},
			},
			res: &common.Resource{
				Name: "test",
			},
			types: common.TypeToResources{
				"test": []*common.Resource{
					{Name: "leased"},
				},
			},
			result: expected{
				err: "fail",
			},
			vmf: &faker{
				fail: true,
			},
		},
		{
			name: "failed cluster creation",
			rc: resourcesConfig{
				"test": {{
					Clusters: []clusterConfig{
						{},
					},
					Vms: []virtualMachineConfig{
						{},
					}},
				},
			},
			res: &common.Resource{
				Name: "test",
			},
			types: common.TypeToResources{
				"test": []*common.Resource{
					{Name: "leased"},
				},
			},
			result: expected{
				err: "fail",
			},
			cf: &faker{
				fail: true,
			},
		},
		{
			name: "running out project",
			rc: resourcesConfig{
				"test": {{
					Clusters: []clusterConfig{
						{},
					},
					Vms: []virtualMachineConfig{
						{},
					}},
				},
			},
			res: &common.Resource{
				Name: "test",
			},
			types: common.TypeToResources{
				"test1": []*common.Resource{
					{Name: "leased"},
				},
			},
			result: expected{
				err: "running out of project while creating resources",
			},
		},
	}
	for _, tc := range testCases {
		defaultClient = &client{
			operationTimeout: time.Second,
			gce:              &fakeVMCreator{f: tc.vmf},
			gke:              &fakeClusterCreator{f: tc.cf},
		}
		ud, err := tc.rc.Construct(tc.res, tc.types)
		if tc.result.err != "" {
			if ud != nil {
				t.Errorf("%s - expected nil user data got %v", tc.name, ud)
			}
			if err == nil || err.Error() != tc.result.err {
				t.Errorf("%s - expected err %s got %v", tc.name, tc.result.err, err)
			}
		} else {
			if err != nil {
				t.Errorf("%s - expected no error got %v", tc.name, err)
			}
			if ud == nil {
				t.Errorf("%s - expected user data got nil", tc.name)
			} else {
				var info ResourceInfo
				if err := ud.Extract(ResourceConfigType, &info); err != nil {
					t.Errorf("%s - unable to parse user data %v", tc.name, err)
				}
				if !reflect.DeepEqual(info, *tc.result.info) {
					t.Errorf("%s - expected info %v got %v instead", tc.name, *tc.result.info, info)
				}
			}
		}
	}
}

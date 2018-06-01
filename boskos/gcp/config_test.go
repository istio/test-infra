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
	"sort"
	"testing"
	"time"

	"k8s.io/test-infra/boskos/common"
	"k8s.io/test-infra/boskos/mason"
	"k8s.io/test-infra/boskos/ranch"
)

func TestParseInvalidConfig(t *testing.T) {
	expected := resourceConfigs{
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
		if !reflect.DeepEqual(expected, *config.(*resourceConfigs)) {
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
				var m mason.Masonable
				m, err = ConfigConverter(config.Config.Content)
				if err != nil {
					t.Errorf("unable to parse config %s %v", config.Name, err)
				}
				needs := common.ResourceNeeds{}
				rc, ok := m.(*resourceConfigs)
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

func (vmc *fakeVMCreator) create(ctx context.Context, p string, c virtualMachineConfig) (*InstanceInfo, error) {
	if vmc.f != nil {
		if err := vmc.f.do(ctx); err != nil {
			return nil, err
		}
	}
	return &InstanceInfo{
		Name: "name",
		Zone: c.Zone,
	}, nil
}

func (vmc *fakeVMCreator) listZones(project string) ([]string, error) {
	return []string{"zone1", "zone2", "zone3"}, nil
}

type fakeClusterCreator struct {
	f *faker
}

func (cc *fakeClusterCreator) create(ctx context.Context, p string, c clusterConfig) (*InstanceInfo, error) {
	if cc.f != nil {
		if err := cc.f.do(ctx); err != nil {
			return nil, err
		}
	}
	return &InstanceInfo{
		Name: "name",
		Zone: c.Zone,
	}, nil
}

func sortInfo(info *ResourceInfo) {
	for _, v := range *info {
		sort.Slice(v.Clusters, func(i, j int) bool { return v.Clusters[i].Zone < v.Clusters[j].Zone })
		sort.Slice(v.VMs, func(i, j int) bool { return v.VMs[i].Zone < v.VMs[j].Zone })
	}
}

func TestResourcesConfig_Construct(t *testing.T) {
	type expected struct {
		err  string
		info *ResourceInfo
	}

	testCases := []struct {
		name      string
		rc        resourceConfigs
		res       *common.Resource
		types     common.TypeToResources
		result    expected
		vmf, cf   *faker
		setClient bool
	}{
		{
			name:      "success",
			setClient: true,
			rc: resourceConfigs{
				"test": {{
					Clusters: []clusterConfig{
						{
							Zone: "specified",
						},
						{},
						{},
					},
					Vms: []virtualMachineConfig{
						{
							Zone: "specified",
						},
						{},
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
						Clusters: []InstanceInfo{
							{
								Name: "name",
								Zone: "specified",
							},
							{
								Name: "name",
								Zone: "zone1",
							},
							{
								Name: "name",
								Zone: "zone2",
							},
						},
						VMs: []InstanceInfo{
							{
								Name: "name",
								Zone: "specified",
							},
							{
								Name: "name",
								Zone: "zone1",
							},
							{
								Name: "name",
								Zone: "zone3",
							},
						},
					},
				},
			},
		},
		{
			name:      "timeout vm creation",
			setClient: true,
			rc: resourceConfigs{
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
			name:      "timeout cluster creation",
			setClient: true,
			rc: resourceConfigs{
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
			name:      "failed vm creation",
			setClient: true,
			rc: resourceConfigs{
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
			name:      "failed cluster creation",
			setClient: true,
			rc: resourceConfigs{
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
			name:      "running out project",
			setClient: true,
			rc: resourceConfigs{
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
		{
			name: "client not set",
			rc: resourceConfigs{
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
				err: "client not set",
			},
		},
	}
	for _, tc := range testCases {
		if tc.setClient {
			c := &Client{
				operationTimeout: time.Second,
				gce:              &fakeVMCreator{f: tc.vmf},
				gke:              &fakeClusterCreator{f: tc.cf},
			}
			SetClient(c)
		} else {
			SetClient(nil)
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
				sortInfo(&info)
				if !reflect.DeepEqual(info, *tc.result.info) {
					t.Errorf("%s - expected info %v got %v instead", tc.name, *tc.result.info, info)
				}
			}
		}
	}
}

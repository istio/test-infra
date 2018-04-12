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
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/container/v1"
)

const (
	defaultSleepTime = 10 * time.Second
	// Defined in https://godoc.org/google.golang.org/api/container/v1#Operation
	operationDone     = "DONE"
	operationAborting = "ABORTING"
)

type clusterConfig struct {
	MachineType           string `json:"machinetype,omitempty"`
	NumNodes              int64  `json:"numnodes,omitempty"`
	Version               string `json:"version,omitempty"`
	Zone                  string `json:"zone,ompitempty"`
	EnableKubernetesAlpha bool   `json:"enablekubernetesalpha"`
}

type containerEngine struct {
	service *container.Service
}

func findVersionMatch(version string, supportedVersion []string) (string, error) {
	for _, v := range supportedVersion {
		if strings.HasPrefix(v, version) {
			return v, nil
		}
	}
	return "", nil
}

func (cc *containerEngine) waitForOperation(ctx context.Context, op *container.Operation, project, zone string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(defaultSleepTime):
			newOp, err := cc.service.Projects.Zones.Operations.Get(project, zone, op.Name).Context(ctx).Do()
			if err != nil {
				return err
			}
			switch newOp.Status {
			case operationDone:
				return nil
			case operationAborting:
				return fmt.Errorf(newOp.StatusMessage)
			default:
				logrus.Infof("operation %s status is %s", newOp.Name, newOp.Status)
			}
		}
	}
}

func (cc *containerEngine) create(ctx context.Context, project string, config clusterConfig) (*InstanceInfo, error) {
	var version string
	name := generateName("gke")
	serverConfig, err := cc.service.Projects.Zones.GetServerconfig(project, config.Zone).Do()
	if err != nil {
		return nil, err
	}
	if config.Version == "" {
		version = serverConfig.DefaultClusterVersion
	} else {
		version, err = findVersionMatch(config.Version, serverConfig.ValidMasterVersions)
		if err != nil {
			return nil, err
		}
	}
	clusterRequest := &container.CreateClusterRequest{
		Cluster: &container.Cluster{
			Name: name,
			InitialClusterVersion: version,
			InitialNodeCount:      config.NumNodes,
			NodeConfig: &container.NodeConfig{
				MachineType: config.MachineType,
			},
			EnableKubernetesAlpha: config.EnableKubernetesAlpha,
		},
	}

	op, err := cc.service.Projects.Zones.Clusters.Create(project, config.Zone, clusterRequest).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	logrus.Infof("Instance %s being created via operation %s, waiting for completion", clusterRequest.Cluster.Name, op.Name)
	if err := cc.waitForOperation(ctx, op, project, config.Zone); err != nil {
		logrus.WithError(err).Errorf("operation %s failed", op.Name)
		return nil, err
	}
	logrus.Infof("Instance %s created via operation %s", clusterRequest.Cluster.Name, op.Name)
	info := &InstanceInfo{Name: name, Zone: config.Zone}

	return info, nil
}

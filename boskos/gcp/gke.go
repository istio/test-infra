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
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	container "google.golang.org/api/container/v1beta1"

	"istio.io/test-infra/toolbox/util"
)

const (
	defaultSleepTime = 10 * time.Second
	readyTimeout     = time.Minute
	// Defined in https://godoc.org/google.golang.org/api/container/v1#Operation
	operationDone     = "DONE"
	operationAborting = "ABORTING"
)

type clusterConfig struct {
	Scopes                  []string                  `json:"scopes,omitempty"`
	MachineType             string                    `json:"machinetype,omitempty"`
	Version                 string                    `json:"version,omitempty"`
	Zone                    string                    `json:"zone,omitempty"`
	NumNodes                int64                     `json:"numnodes,omitempty"`
	ReleaseChannel          *container.ReleaseChannel `json:"releasechannel,omitempty"`
	NetworkPolicy           *container.NetworkPolicy  `json:"networkpolicy,omitempty"`
	EnableKubernetesAlpha   bool                      `json:"enablekubernetesalpha"`
	EnableWorkloadIdentity  bool                      `json:"enableworkloadidentity"`
	EnableClientCertificate bool                      `json:"enableclientcertificate"`
}

type containerEngine struct {
	service *container.Service
}

func findVersionMatch(version string, supportedVersion []string) string {
	for _, v := range supportedVersion {
		if strings.HasPrefix(v, version) {
			return v
		}
	}
	return ""
}

func checkCluster(kubeconfig string) error {
	_, err := util.Shell("kubectl --kubeconfig=%s get ns", kubeconfig)
	return err
}

func (cc *containerEngine) waitForReady(ctx context.Context, cluster, project, zone string) error {
	logrus.Infof("Verifying that cluster %s in zone %s for project %s is ready", cluster, zone, project)
	kubeconfigFile, err := ioutil.TempFile("", "kubeconfig")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(kubeconfigFile.Name()); err != nil {
			logrus.WithError(err).Errorf("failed to delete file %s", kubeconfigFile.Name())
		}

	}()

	if err := SetKubeConfig(project, zone, cluster, kubeconfigFile.Name()); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(defaultSleepTime):
			err := checkCluster(kubeconfigFile.Name())
			if err != nil {
				logrus.WithError(err).Errorf("cluster %s in zone %s for project %s is not ready", cluster, zone, project)

			} else {
				logrus.Infof("cluster %s in zone %s for project %s is ready", cluster, zone, project)
				return nil
			}
		}
	}
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

	// If release channel is specified, find the corresponding valid version in the release channel.
	if config.ReleaseChannel != nil && config.ReleaseChannel.Channel != "" {
		var chConfig *container.ReleaseChannelConfig
		for _, ch := range serverConfig.Channels {
			if ch.Channel == config.ReleaseChannel.Channel {
				chConfig = ch
			}
		}

		if chConfig == nil {
			return nil, fmt.Errorf("invalid release channel from the config file: %s", config.ReleaseChannel.Channel)
		}

		if config.Version == "" {
			version = chConfig.DefaultVersion
		} else {
			validVersions := make([]string, len(chConfig.AvailableVersions))
			for i, v := range chConfig.AvailableVersions {
				validVersions[i] = v.Version
			}
			version = findVersionMatch(config.Version, validVersions)
		}
	} else {
		if config.Version == "" {
			version = serverConfig.DefaultClusterVersion
		} else {
			version = findVersionMatch(config.Version, serverConfig.ValidMasterVersions)
		}
	}

	clusterRequest := &container.CreateClusterRequest{
		Cluster: &container.Cluster{
			ReleaseChannel:        config.ReleaseChannel,
			Name:                  name,
			InitialClusterVersion: version,
			InitialNodeCount:      config.NumNodes,
			NodeConfig: &container.NodeConfig{
				MachineType: config.MachineType,
				OauthScopes: config.Scopes,
			},
			NetworkPolicy:         config.NetworkPolicy,
			EnableKubernetesAlpha: config.EnableKubernetesAlpha,
		},
	}
	// Since Boskos can pick any project in the pool, we need to make sure the identity namespace ties
	// to the correct project id.
	if config.EnableWorkloadIdentity {
		clusterRequest.Cluster.WorkloadIdentityConfig = &container.WorkloadIdentityConfig{
			IdentityNamespace: fmt.Sprintf("%s.svc.id.goog", project),
		}
	}
	if config.EnableClientCertificate {
		clusterRequest.Cluster.MasterAuth.ClientCertificateConfig.IssueClientCertificate = true
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
	readyCtx, cancel := context.WithTimeout(ctx, readyTimeout)
	defer cancel()
	if err := cc.waitForReady(readyCtx, name, project, config.Zone); err != nil {
		logrus.WithError(err).Errorf("cluster %s in zone %s for project %s is not usable", name, config.Zone, project)
		return nil, err
	}
	return info, nil
}

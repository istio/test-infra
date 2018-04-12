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
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
	"gopkg.in/yaml.v2"
	"k8s.io/test-infra/boskos/common"
	"k8s.io/test-infra/boskos/mason"
)

var (
	seededRand     = rand.New(rand.NewSource(time.Now().UnixNano()))
	serviceAccount = flag.String("service-account", "", "Path to projects service account")
)

const (
	// ResourceConfigType defines the GCP config type
	ResourceConfigType      = "GCPResourceConfig"
	defaultOperationTimeout = 20 * time.Minute
	charset                 = "abcdefghijklmnopqrstuvwxyz1234567890"
)

type projectConfig struct {
	Type     string                 `json:"type,omitempty"`
	Clusters []clusterConfig        `json:"clusters,omitempty"`
	Vms      []virtualMachineConfig `json:"vms,omitempty"`
}

type resourcesConfig struct {
	ProjectConfigs []projectConfig `json:"projectconfigs,omitempty"`
}

type instanceInfo struct {
	Name string `json:"name"`
	Zone string `json:"zone"`
}

type projectInfo struct {
	Name     string         `json:"name"`
	Clusters []instanceInfo `json:"clusters,omitempty"`
	VMs      []instanceInfo `json:"vms,omitempty"`
}

// ResourceInfo holds information about the resource created, such that it can used
type ResourceInfo struct {
	ProjectsInfo []projectInfo `json:"projectsinfo,omitempty"`
}

type vmCreator interface {
	create(context.Context, string, virtualMachineConfig) (*instanceInfo, error)
}

type clusterCreator interface {
	create(context.Context, string, clusterConfig) (*instanceInfo, error)
}

type client struct {
	gke clusterCreator
	gce vmCreator
	operationTimeout time.Duration
}

// Construct implements Masonable interface
func (rc *resourcesConfig) Construct(res *common.Resource, types common.TypeToResources) (common.UserData, error) {
	info := ResourceInfo{}
	var err error

	gcpClient, err := newClient()
	if err != nil {
		return nil, err
	}
	// Copy
	typesCopy := types

	popProject := func(rType string) *common.Resource {
		if len(typesCopy[rType]) == 0 {
			return nil
		}
		r := typesCopy[rType][len(typesCopy[rType])-1]
		typesCopy[rType] = typesCopy[rType][:len(typesCopy[rType])-1]
		return r
	}

	ctx, cancel := context.WithTimeout(context.Background(), gcpClient.operationTimeout)
	defer cancel()
	errGroup, derivedCtx := errgroup.WithContext(ctx)
	var infoMutex sync.RWMutex

	// Here we know that resources are of project type
	for _, pc := range rc.ProjectConfigs {
		project := popProject(pc.Type)
		if project == nil {
			err = fmt.Errorf("running out of project while creating resources")
			logrus.WithError(err).Errorf("unable to create resources")
			return nil, err
		}
		pi := projectInfo{Name: project.Name}
		for _, cl := range pc.Clusters {
			errGroup.Go(func() error {
				clusterInfo, err := gcpClient.gke.create(derivedCtx, project.Name, cl)
				if err != nil {
					logrus.WithError(err).Errorf("unable to create cluster on project %s", project.Name)
					return err
				}
				infoMutex.Lock()
				pi.Clusters = append(pi.Clusters, *clusterInfo)
				infoMutex.Unlock()
				return nil
			})
		}
		for _, vm := range pc.Vms {
			errGroup.Go(func() error {
				vmInfo, err := gcpClient.gce.create(derivedCtx, project.Name, vm)
				if err != nil {
					logrus.WithError(err).Errorf("unable to create vm on project %s", project.Name)
					return err
				}
				infoMutex.Lock()
				pi.VMs = append(pi.VMs, *vmInfo)
				infoMutex.Unlock()
				return nil
			})
		}
		infoMutex.Lock()
		info.ProjectsInfo = append(info.ProjectsInfo, pi)
		infoMutex.Unlock()
	}

	if err := errGroup.Wait(); err != nil {
		logrus.WithError(err).Errorf("failed to construct resources for %s", res.Name)
		return nil, err
	}

	userData := common.UserData{}
	if err := userData.Set(ResourceConfigType, &info); err != nil {
		logrus.WithError(err).Errorf("unable to set %s user data", ResourceConfigType)
		return nil, err
	}
	return userData, nil
}

// ConfigConverter implements mason.ConfigConverter
func ConfigConverter(in string) (mason.Masonable, error) {
	var config resourcesConfig
	if err := yaml.Unmarshal([]byte(in), &config); err != nil {
		logrus.WithError(err).Errorf("unable to parse %s", in)
		return nil, err
	}
	return &config, nil
}

func newClient() (*client, error) {
	var (
		oauthClient *http.Client
		err         error
	)
	if *serviceAccount != "" {
		var data []byte
		data, err = ioutil.ReadFile(*serviceAccount)
		if err != nil {
			return nil, err
		}
		var conf *jwt.Config
		conf, err = google.JWTConfigFromJSON(data, compute.CloudPlatformScope)
		if err != nil {
			return nil, err
		}
		oauthClient = conf.Client(context.Background())
	} else {
		oauthClient, err = google.DefaultClient(context.Background(), compute.CloudPlatformScope)
		if err != nil {
			return nil, err
		}
	}
	gkeService, err := container.New(oauthClient)
	if err != nil {
		return nil, err
	}
	gceService, err := compute.New(oauthClient)
	if err != nil {
		return nil, err
	}
	return &client{
		gke:              &containerEngine{gkeService},
		gce:              &computeEngine{gceService},
		operationTimeout: defaultOperationTimeout,
	}, nil
}

func randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func generateName(prefix string) string {
	date := time.Now().Format("010206")
	randString := randomString(10)
	return fmt.Sprintf("%s-%s-%s", prefix, date, randString)
}

// Install kubeconfig for a given resource. It will create only one file with all contexts.
func (r ResourceInfo) Install(kubeconfig string) error {
	for _, p := range r.ProjectsInfo {
		for _, c := range p.Clusters {
			if err := SetKubeConfig(p.Name, c.Zone, c.Name, kubeconfig); err != nil {
				logrus.WithError(err).Errorf("failed to set kubeconfig")
				return err
			}
		}
	}
	return nil
}

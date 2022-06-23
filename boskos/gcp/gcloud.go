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
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	container "google.golang.org/api/container/v1beta1"
	clientapi "k8s.io/client-go/tools/clientcmd/api/v1"
)

// SetKubeConfig saves kube config from a given cluster to the given location
// It uses client certificate if it presents.
func SetKubeConfig(project, zone, cluster, kubeconfig string) error {
	if err := os.Setenv("KUBECONFIG", kubeconfig); err != nil {
		return err
	}

	clusterJSON, err := ShellSilent(
		"gcloud container clusters describe %s --project=%s --zone=%s --format=json",
		cluster, project, zone)
	if err != nil {
		return err
	}

	clusterObj := container.Cluster{}
	if err = json.Unmarshal([]byte(clusterJSON), &clusterObj); err != nil {
		return err
	}

	if clusterObj.MasterAuth == nil ||
		(len(clusterObj.MasterAuth.ClientCertificate) == 0 && len(clusterObj.MasterAuth.ClientKey) == 0) {
		_, err := ShellSilent(
			"gcloud container clusters get-credentials %s --project=%s --zone=%s",
			cluster, project, zone)
		return err
	}

	ca, err := base64.StdEncoding.DecodeString(clusterObj.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return err
	}
	clientCert, err := base64.StdEncoding.DecodeString(clusterObj.MasterAuth.ClientCertificate)
	if err != nil {
		return err
	}
	clientKey, err := base64.StdEncoding.DecodeString(clusterObj.MasterAuth.ClientKey)
	if err != nil {
		return err
	}

	config := clientapi.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: []clientapi.NamedCluster{
			{
				Name: cluster,
				Cluster: clientapi.Cluster{
					Server:                   "https://" + clusterObj.Endpoint,
					CertificateAuthorityData: ca,
				},
			},
		},
		AuthInfos: []clientapi.NamedAuthInfo{
			{
				Name: cluster,
				AuthInfo: clientapi.AuthInfo{
					ClientCertificateData: clientCert,
					ClientKeyData:         clientKey,
				},
			},
		},
		Contexts: []clientapi.NamedContext{
			{
				Name: cluster,
				Context: clientapi.Context{
					Cluster:  cluster,
					AuthInfo: cluster,
				},
			},
		},
		CurrentContext: cluster,
	}

	kubeconfigData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(kubeconfig, kubeconfigData, 0o666)
}

// ActivateServiceAccount activates a service account for gcloud
func ActivateServiceAccount(serviceAccount string) error {
	_, err := ShellSilent(
		"gcloud auth activate-service-account --key-file=%s",
		serviceAccount)
	return err
}

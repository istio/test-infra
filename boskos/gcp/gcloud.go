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
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

func runCommand(name string, args ...string) error {
	logrus.Infof("running command %s %s", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		logrus.Infof("failed to start command %s %s", name, strings.Join(args, " "))
		return err
	}
	return cmd.Wait()
}

// SetKubeConfig saves kube config from a given cluster to the given location
func SetKubeConfig(project, zone, cluster, kubeconfig string) error {
	if err := os.Setenv("KUBECONFIG", kubeconfig); err != nil {
		return err
	}
	return runCommand("gcloud", "container", "clusters", "get-credentials", cluster,
		"--project", project, "--zone", zone)
}

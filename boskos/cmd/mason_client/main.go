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

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"k8s.io/test-infra/boskos/client"
	"k8s.io/test-infra/boskos/common"
	"k8s.io/test-infra/boskos/mason"

	"istio.io/test-infra/boskos/gcp"
)

const (
	defaultSleepTime = 10 * time.Second
	defaultTimeout   = "10m"
)

func defaultKubeconfig() string {
	home := os.Getenv("HOME")
	if home == "" {
		return ""
	}
	return fmt.Sprintf("%s/.kube/config", home)
}

var (
	owner       = flag.String("owner", "", "")
	rType       = flag.String("type", "", "Type of resource to acquire")
	timeoutStr  = flag.String("timeout", defaultTimeout, "Timeout ")
	kubecfgPath = flag.String("kubeconfig-save", defaultKubeconfig(), "Path to write kubeconfig file to")
	infoSave    = flag.String("info-save", "", "Path to save info")
)

type masonClient struct {
	mason *mason.Client
	wg    sync.WaitGroup
}

func (m *masonClient) acquire(ctx context.Context, rtype, state string) (*common.Resource, error) {
	for {
		select {
		case <-time.After(defaultSleepTime):
			logrus.Infof("Attempting to acquire resource")
			res, err := m.mason.Acquire(rtype, common.Free, state)
			if err == nil {
				logrus.Infof("Resource %s acquired", res.Name)
				return res, nil
			}
			logrus.Infof("Failed to acquire resource")
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (m *masonClient) release(res common.Resource) {
	if err := m.mason.ReleaseOne(res.Name, common.Dirty); err != nil {
		logrus.WithError(err).Warningf("unable to release resource %s", res.Name)
		return
	}
	logrus.Infof("Released resource %s", res.Name)
}

func (m *masonClient) update(ctx context.Context, state string) {
	updateTick := time.NewTicker(defaultSleepTime).C
	go func() {
		for {
			select {
			case <-updateTick:
				if err := m.mason.UpdateAll(state); err != nil {
					logrus.WithError(err).Warningf("unable to update resources to state %s", state)
				}
				logrus.Infof("Updated resources")
			case <-ctx.Done():
				m.wg.Done()
				return
			}
		}
	}()
	m.wg.Add(1)
}

func saveUserdataToFile(ud common.UserData, key, path string) error {
	v, ok := ud[key]
	if !ok {
		return nil
	}
	return ioutil.WriteFile(path, []byte(v), 0644)
}

func wait() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}

func main() {
	flag.Parse()
	if *rType == "" {
		logrus.Errorf("flag --type must be set")
		flag.Usage()
		return
	}
	if *owner == "" {
		logrus.Errorf("flag --owner must be set")
		flag.Usage()
		return
	}
	timeout, err := time.ParseDuration(*timeoutStr)
	if err != nil {
		logrus.Errorf("unable to parse --timeout %s", *timeoutStr)
		flag.Usage()
		return
	}
	client := masonClient{mason: mason.NewClient(client.NewClient(*owner, *client.BoskosURL))}
	if *kubecfgPath == "" {
		logrus.Panic("flag --type must be set")
	}
	c1, acquireCancel := context.WithTimeout(context.Background(), timeout)
	defer acquireCancel()
	res, err := client.acquire(c1, *rType, common.Busy)
	if err != nil {
		logrus.WithError(err).Panicf("unable to find a resource")
	}
	defer client.release(*res)
	c2, updateCancel := context.WithCancel(context.Background())
	defer updateCancel()
	client.update(c2, common.Busy)

	for cType := range res.UserData {
		switch cType {
		case gcp.ResourceConfigType:
			if *kubecfgPath != "" {
				var info gcp.ResourceInfo
				if err := res.UserData.Extract(gcp.ResourceConfigType, &info); err != nil {
					logrus.WithError(err).Panicf("unable to parse %s", gcp.ResourceConfigType)
				}
				if err := info.Install(*kubecfgPath); err != nil {
					logrus.WithError(err).Panicf("unable to install %s", gcp.ResourceConfigType)
				}
			}
			if *infoSave != "" {
				if err := saveUserdataToFile(res.UserData, gcp.ResourceConfigType, *infoSave); err != nil {
					logrus.WithError(err).Panicf("unable to save info to %s", *infoSave)
				}
				logrus.Infof("Saved user data to %s", *infoSave)
			}
			break
		}
	}
	logrus.Infof("READY")
	logrus.Infof("Type CTRL-C to interrupt")
	wait()
	updateCancel()
	client.wg.Wait()
}

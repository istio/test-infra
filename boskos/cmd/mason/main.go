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
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"istio.io/test-infra/boskos/gcp"

	"k8s.io/test-infra/boskos/mason"
)

const (
	defaultUpdatePeriod = time.Minute * 1
)

var (
	configPath = flag.String("config", "", "Path to persistent volume to load configs")
)

func main() {
	flag.Parse()
	logrus.SetFormatter(&logrus.JSONFormatter{})

	if *configPath == "" {
		logrus.Panic("--config must be set")
	}

	mason := mason.NewMasonFromFlags()

	// Registering Masonable Converters
	if err := mason.RegisterConfigConverter(gcp.ResourceConfigType, gcp.ConfigConverter); err != nil {
		logrus.WithError(err).Panicf("unable tp register config converter")
	}
	if err := mason.UpdateConfigs(*configPath); err != nil {
		logrus.WithError(err).Panicf("failed to update mason config")
	}
	go func() {
		for range time.NewTicker(defaultUpdatePeriod).C {
			if err := mason.UpdateConfigs(*configPath); err != nil {
				logrus.WithError(err).Warning("failed to update mason config")
			}
		}
	}()

	mason.Start()
	defer mason.Stop()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}

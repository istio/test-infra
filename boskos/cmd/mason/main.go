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

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/boskos/client"
	"k8s.io/test-infra/boskos/mason"

	"istio.io/test-infra/boskos/gcp"
)

const (
	defaultUpdatePeriod      = time.Minute
	defaultChannelSize       = 15
	defaultCleanerCount      = 15
	defaultBoskosRetryPeriod = 15 * time.Second
	defaultOwner             = "mason"
)

var (
	boskosURL         = flag.String("boskos-url", "http://boskos", "Boskos Server URL")
	channelBufferSize = flag.Int("channel-buffer-size", defaultChannelSize, "Channel Size")
	cleanerCount      = flag.Int("cleaner-count", defaultCleanerCount, "Number of threads running cleanup")
	configPath        = flag.String("config", "", "Path to persistent volume to load configs")
	serviceAccount    = flag.String("service-account", "", "Path to projects service account")
)

func main() {
	flag.Parse()
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if *configPath == "" {
		logrus.Fatalf("--config must be set")
	}
	if *serviceAccount != "" {
		if err := gcp.ActivateServiceAccount(*serviceAccount); err != nil {
			logrus.WithError(err).Fatal("cannot activate service account")
		}
	}

	client := client.NewClient(defaultOwner, *boskosURL)
	gcpClient, err := gcp.NewClient(*serviceAccount)
	if err != nil {
		logrus.WithError(err).Fatal("unable to create gcp client")
	}
	gcp.SetClient(gcpClient)

	mason := mason.NewMason(*channelBufferSize, *cleanerCount, client, defaultBoskosRetryPeriod)

	// Registering Masonable Converters
	if err := mason.RegisterConfigConverter(gcp.ResourceConfigType, gcp.ConfigConverter); err != nil {
		logrus.WithError(err).Fatalf("unable tp register config converter")
	}
	if err := mason.UpdateConfigs(*configPath); err != nil {
		logrus.WithError(err).Fatalf("failed to update mason config")
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

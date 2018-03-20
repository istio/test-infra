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
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/context"
	"google.golang.org/api/option"

	"istio.io/test-infra/toolbox/metrics"
	"istio.io/test-infra/toolbox/metrics/coverage"
)

const (
	defaultUpdateInterval = 2 * time.Microsecond
	defaultUpdateTimeout  = 30 * time.Second
)

var (
	port               = flag.Int("listen_port", 9103, "Port on which to expose metrics and web interface.")
	gcsBucket          = flag.String("bucket", "istio-code-coverage", "GCS bucket name.")
	codeCovTrackJob    = flag.String("coverage_job", "codecov_master", "In which job we are tracking code coverage.")
	githubRepo         = flag.String("repo", "istio", "Repo for which we are capturing code coverage")
	serviceAccountJSON = flag.String("service_account_json", "", "Path to the service account JSON")

	ms *metrics.Publisher
	// Cannot set storage in init since we need flags.
	storage = coverage.NewGCSStorage()
)

func newMetricPublisher() *metrics.Publisher {
	suite := metrics.Suite{
		"codecov": coverage.NewMetric(storage),
	}
	return metrics.NewPublisher(suite, defaultUpdateInterval, defaultUpdateTimeout)
}

func init() {
	ms = newMetricPublisher()
	ms.RegisterMetrics()
}

func main() {
	flag.Parse()
	var options []option.ClientOption
	if *serviceAccountJSON != "" {
		options = append(options, option.WithCredentialsFile(*serviceAccountJSON))
	}

	if err := storage.Set(*gcsBucket, *githubRepo, *codeCovTrackJob, options); err != nil {
		glog.Fatalf("unable to create storage %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := ms.Publish(ctx); err != nil {
			glog.Fatal(err)
		}
	}()
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
			glog.Fatal(err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	cancel()
}

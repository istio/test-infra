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

package flakes

import (
	"context"
	"encoding/json"
	"strconv"

	u "istio.io/test-infra/toolbox/util"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	bucket             = "k8s-metrics"
	latestFlakesMetric = "istio-job-flakes-latest.json"
)

var (
	gcsClient = u.NewGCSClient(bucket)
)

// FlakeGauge implement the metrics.Metric interface
type FlakeGauge struct {
	gauge *prometheus.GaugeVec
}

// FlakeMetric is how the metric is defined in the json output by bigquery metrics
type FlakeMetric struct {
	Consistency string `json:"consistency"`
	Job         string `json:"job"`
	Passed      string `json:"passed"`
	Runs        string `json:"runs"`
	Stamp       string `json:"stamp"`
}

// NewMetric instantiates a new Flake metric
func NewMetric() *FlakeGauge {
	return &FlakeGauge{
		gauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "consistency",
				Help: "In past 7 days, Number of successful runs over number of total runs",
			},
			[]string{"job"},
		),
	}
}

// GetCollector implements metrics.Metric interface
func (f *FlakeGauge) GetCollector() prometheus.Collector {
	return f.gauge
}

func (f *FlakeGauge) update() error {
	flattened, err := gcsClient.Read(latestFlakesMetric)
	if err != nil {
		return err
	}
	var flakes []FlakeMetric
	if err = json.Unmarshal([]byte(flattened), &flakes); err != nil {
		return err
	}
	for _, flake := range flakes {
		glog.Infof("setting %s to %s", flake.Job, flake.Consistency)
		if consistency, err := strconv.ParseFloat(flake.Consistency, 64); err != nil {
			glog.Errorf("Failed to convert %s to float64: %v", flake.Consistency, err)
		} else {
			f.gauge.WithLabelValues(flake.Job).Set(consistency)
		}
	}
	return nil
}

// Update implements metrics.Metric interface
func (f *FlakeGauge) Update(ctx context.Context) error {
	glog.Infof("upading flakes matrics")
	errc := make(chan error)
	go func() {
		errc <- f.update()
	}()
	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

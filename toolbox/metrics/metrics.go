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

package metrics

import (
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
)

// Suite of metrics to collect from.
type Suite map[string]Metric

// NewPublisher creates a publisher from a suite and update interval and timeout
func NewPublisher(s Suite, updateInterval, updateTimeout time.Duration) *Publisher {
	return &Publisher{
		suite:          s,
		updateInterval: updateInterval,
		updateTimeout:  updateTimeout,
	}
}

// Publisher manages publication of multiple metrics
type Publisher struct {
	suite                         Suite
	updateInterval, updateTimeout time.Duration
}

// Metric is used to collect data
type Metric interface {
	Update(ctx context.Context) error
	GetCollector() prometheus.Collector
}

// RegisterMetrics should be called in init
func (p *Publisher) RegisterMetrics() {
	for _, m := range p.suite {
		prometheus.MustRegister(m.GetCollector())
	}
}

// Update will update each metrics in the suite
func (p *Publisher) Update(ctx context.Context) {
	nCtx, cancel := context.WithTimeout(ctx, p.updateTimeout)
	defer cancel()
	for k, m := range p.suite {
		glog.Infof("Updating metric %s", k)
		if err := m.Update(nCtx); err != nil {
			glog.Warning("failed to update metric for %s. %v", k, err)
		}

	}
}

// Publish will update Prometheus HTTP Server
func (p *Publisher) Publish(ctx context.Context) error {
	glog.Infof("Starting publishing thread")
	defer glog.Infof("Terminating publishing thread")
	tick := time.Tick(p.updateInterval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick:
			p.Update(ctx)
		}
	}
	return nil
}

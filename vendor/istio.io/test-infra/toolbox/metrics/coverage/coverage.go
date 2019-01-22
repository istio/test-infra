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

package coverage

import (
	"bufio"
	"context"
	"io"
	"regexp"
	"strconv"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

// Storage abstraction for testing
type Storage interface {
	GetLatest(ctx context.Context) (io.ReadCloser, error)
	GetLabel() string
}

// Metric implement the metrics.Metric interface
type Metric struct {
	coverage *prometheus.GaugeVec
	storage  Storage
}

// Coverage holds go package code coverage percentage
type Coverage map[string]float64

// NewMetric instantiates a new Coverage metric
func NewMetric(s Storage) *Metric {
	return &Metric{
		coverage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "package_code_coverage",
				Help: "Package code coverage.",
			},
			[]string{"package", "repo"},
		),
		storage: s,
	}
}

// GetCollector implements Metric interface
func (m *Metric) GetCollector() prometheus.Collector {
	return m.coverage
}

// Update implements Metric interface
func (m *Metric) Update(ctx context.Context) error {
	r, err := m.storage.GetLatest(ctx)
	if err != nil {
		return err
	}
	coverage, err := getCoverage(r)
	if err != nil {
		return err
	}
	for pkg, percent := range coverage {
		m.coverage.WithLabelValues(pkg, m.storage.GetLabel()).Set(percent)
	}
	return nil
}

func getCoverage(r io.ReadCloser) (Coverage, error) {
	cov := Coverage{}
	defer func() {
		if err := r.Close(); err != nil {
			glog.Errorf("unable to close file %v", err)
		}
	}()

	//Line example: "istio.io/mixer/adapter/denyChecker	99"
	scanner := bufio.NewScanner(r)
	reg := regexp.MustCompile(`(.*)\t(.*)`)
	for scanner.Scan() {
		if m := reg.FindStringSubmatch(scanner.Text()); len(m) == 3 {
			if n, err := strconv.ParseFloat(m[2], 64); err != nil {
				glog.Errorf("Failed to parse codecov file: %s, %v", scanner.Text(), err)
			} else {
				cov[m[1]] = n
			}
		} else {
			glog.Errorf("Failed to parse codecov file: %s, broken line", scanner.Text())
		}
	}
	return cov, scanner.Err()
}

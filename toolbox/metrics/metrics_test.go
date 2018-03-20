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

package metrics

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	m *Publisher
)

func newGauge(n string) prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Name: n,
		Help: fmt.Sprintf("help for %s", n),
	})
}

func init() {
	s := Suite{
		"test1": &fakeMetric{m: newGauge("test1")},
	}
	m = NewPublisher(s, time.Millisecond, time.Millisecond)
	m.RegisterMetrics()
}

type fakeMetric struct {
	m     prometheus.Gauge
	count float64
	lock  sync.RWMutex
}

func (f *fakeMetric) Update(ctx context.Context) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.m.Set(f.count)
	f.count++
	return nil
}

func (f *fakeMetric) GetCollector() prometheus.Collector {
	return f.m
}

func TestPublisher_Update(t *testing.T) {
	c, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	m.Update(c)
	if c.Err() != nil {
		t.Error(c.Err())
	}
	f := m.suite["test1"].(*fakeMetric)
	if f.count != 1.0 {
		t.Error("counter should be incremented")
	}
}

func TestPublisher_Publish(t *testing.T) {
	c, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	go func() {
		if err := m.Publish(c); err != c.Err() {
			t.Errorf("Error should match %v received %v instead", c.Err(), err)
		}
	}()
	select {
	case <-time.After(2 * time.Second):
		t.Error("test timed out, context should have finished")
		t.FailNow()
	case <-c.Done():
	}
	f := m.suite["test1"].(*fakeMetric)
	if f.count < 100.0 {
		t.Error("counter should be incremented")
	}
}

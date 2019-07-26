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
	"bytes"
	"context"
	"io"
	"testing"
	"time"
)

const (
	fakeData = `istio.io/istio/mixer/adapter/opa	80.40
istio.io/istio/mixer/pkg/attribute	99.20
istio.io/istio/mixer/pkg/runtime2/routing	100.00
istio.io/istio/mixer/pkg/server	99.40`
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

type fakeStorage struct {
	data string
}

func (f *fakeStorage) GetLatest(ctx context.Context) (io.ReadCloser, error) {
	return nopCloser{bytes.NewBufferString(f.data)}, nil
}

func (f *fakeStorage) GetLabel() string {
	return "master"
}

func TestMetric_Update(t *testing.T) {
	m := NewMetric(&fakeStorage{data: fakeData})

	c, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	errc := make(chan error)
	go func() {
		errc <- m.Update(c)
	}()
	select {
	case err := <-errc:
		if err != nil {
			t.Errorf("Unexpected error %v", err)
			t.FailNow()
		}
	case <-time.After(2 * time.Second):
		t.Errorf("timed out")
		t.FailNow()
	case <-c.Done():
	}
	after := m.coverage.WithLabelValues("istio.io/istio/mixer/adapter/opa", m.storage.GetLabel())
	if after.Desc() == nil {
		t.Errorf("%v should not be nil", after.Desc())
	}
}

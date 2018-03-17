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

package sisyphus

import (
	"log"
)

// Istio uses Kettle from k8s tooling to export data to BigQuery.
// This interface enables additional storage needs for clients besides istio
// and facilitates mocking in tests.
type ISisyphusStorage interface {
	Store(jobName, sha string, newFlakeStat FlakeStat) error
}

type SisyphusStorage struct{}

func NewSisyphusStorage() *SisyphusStorage {
	return &SisyphusStorage{}
}

// Store records FlakeStat to durable storage
func (s *SisyphusStorage) Store(jobName, sha string, newFlakeStat FlakeStat) error {
	log.Printf("newFlakeStat = %v\n", newFlakeStat)
	return nil
}

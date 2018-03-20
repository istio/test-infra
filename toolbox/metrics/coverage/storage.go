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

package coverage

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/golang/glog"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GCSStorage implements the Storage interface
type GCSStorage struct {
	bucketHandle  *storage.BucketHandle
	repo, jobName string
	latest        chan string
}

// NewGCSStorage instantiates a new GCSStorage
func NewGCSStorage() *GCSStorage {
	return &GCSStorage{
		latest: make(chan string, 10),
	}
}

// Set needs be called in the main, as the GCSStorage is created in the binary init function per Prometheus Requirement
func (g *GCSStorage) Set(bucket, repo, jobName string, options []option.ClientOption) error {
	storageClient, err := storage.NewClient(context.Background(), options...)
	if err != nil {
		return err
	}
	g.bucketHandle = storageClient.Bucket(bucket)
	g.repo = repo
	g.jobName = jobName
	return nil
}

// Note that I tried using PubSub notification, but it appears that those appear broken for golang.
// https://github.com/GoogleCloudPlatform/google-cloud-node/issues/1502
// Instead we're looking for all the files, and taking the last created.
func (g *GCSStorage) getLatest(ctx context.Context) error {
	q := &storage.Query{
		Prefix:   fmt.Sprintf("%s/%s/", g.repo, g.jobName),
		Versions: false,
	}
	i := g.bucketHandle.Objects(ctx, q)
	var latest string
	latestCreated := time.Time{}

	for {
		objAttrs, err := i.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if objAttrs.Created.After(latestCreated) {
			latest = objAttrs.Name
			latestCreated = objAttrs.Created
		}
	}
	glog.Infof("Found latest %s", latest)
	g.latest <- latest
	return nil
}

// GetLabel returns the repo where coverage was gathered.
func (g *GCSStorage) GetLabel() string {
	return g.repo
}

// GetLatest returns an io.ReadCloser for the bucket file found.
func (g *GCSStorage) GetLatest(ctx context.Context) (io.ReadCloser, error) {
	var latest string
	errc := make(chan error)
	go func() {
		errc <- g.getLatest(ctx)
	}()
	select {
	case err := <-errc:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case latest = <-g.latest:
		glog.Infof("Received latest %s", latest)
	}
	return g.bucketHandle.Object(latest).NewReader(ctx)
}

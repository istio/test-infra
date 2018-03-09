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

package util

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
)

// GCSClient masks RPCs to gcs as local procedures
type GCSClient struct {
	client *storage.Client
	bucket string
}

// NewGCSClient creates a new GCSClient
func NewGCSClient(bucket string) *GCSClient {
	gcsClient, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create a gcs client, %v\n", err)
		return nil
	}
	return &GCSClient{
		client: gcsClient,
		bucket: bucket,
	}
}

// GetReaderOnFile return a GCS reader on the requested obj
// Caller is responsible to close reader afterwards.
func (gcs *GCSClient) GetReaderOnFile(obj string) (*storage.Reader, error) {
	ctx := context.Background()
	r, err := gcs.client.Bucket(gcs.bucket).Object(obj).NewReader(ctx)
	if err != nil {
		log.Printf("Failed to download file %s/%s from gcs, %v\n", gcs.bucket, obj, err)
		return nil, err
	}
	return r, nil
}

// Read gets a file and return a string
func (gcs *GCSClient) Read(obj string) (string, error) {
	r, err := gcs.GetReaderOnFile(obj)
	if err != nil {
		return "", err
	}
	defer func() {
		if err = r.Close(); err != nil {
			log.Printf("Failed to close gcs file reader, %v\n", err)
		}
	}()
	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(r); err != nil {
		log.Printf("Failed to read from gcs reader, %v\n", err)
		return "", err
	}
	return buf.String(), nil
}

// Write writes text to file on gcs
func (gcs *GCSClient) Write(obj, txt string) error {
	ctx := context.Background()
	w := gcs.client.Bucket(gcs.bucket).Object(obj).NewWriter(ctx)
	if _, err := fmt.Fprintf(w, txt); err != nil {
		log.Printf("Failed to write to gcs: %v\n", err)
	}
	return w.Close()
}

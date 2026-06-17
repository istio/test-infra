# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

repo_root = $(shell git rev-parse --show-toplevel)

lint: lint-all

fmt: format-go tidy-go

test:
	@go test -race ./...
	@(cd tools/prowgen; go test -race ./...)
	@(cd authentikos; go test -race ./...)

gen: generate-config fmt mirror-licenses

gen-check: gen check-clean-repo

# Generate the canonical (GKE) job config. Jobs are cloud-agnostic; this tree is the source of truth.
generate-config:
	@rm -fr prow/cluster/gke/jobs/*/*/*.gen.yaml
	@(cd tools/prowgen/cmd/prowgen; go run main.go --input-dir=$(repo_root)/prow/config/jobs --output-dir=$(repo_root)/prow/cluster/gke/jobs write)
	@go run tools/prowtrans/cmd/prowtrans/main.go --requirement-presets=./prow/config/jobs/.base.yaml --configs=./prow/config/istio-private_jobs --input=./prow/config/jobs
	@go run tools/prowtrans/cmd/prowtrans/main.go --requirement-presets=./prow/config/jobs/.base.yaml --configs=./prow/config/experimental --input=./prow/config/jobs

# Mirror the canonical job config into the EKS (AWS) cluster tree consumed by the -aws deploy targets.
generate-config-aws: generate-config
	@rm -fr prow/cluster/eks/jobs && cp -a prow/cluster/gke/jobs prow/cluster/eks/jobs

diff-config:
	@(cd tools/prowgen/cmd/prowgen; GOARCH=$(GOARCH) GOOS=$(GOOS) go run main.go --input-dir=$(repo_root)/prow/config/jobs --output-dir=$(repo_root)/prow/cluster/gke/jobs diff)

diff-config-aws:
	@(cd tools/prowgen/cmd/prowgen; GOARCH=$(GOARCH) GOOS=$(GOOS) go run main.go --input-dir=$(repo_root)/prow/config/jobs --output-dir=$(repo_root)/prow/cluster/eks/jobs diff)

include common/Makefile.common.mk

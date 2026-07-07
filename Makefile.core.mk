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
	@rm -fr prow/gcp/cluster/jobs/*/*/*.gen.yaml
	@(cd tools/prowgen/cmd/prowgen; go run main.go --input-dir=$(repo_root)/prow/gcp/config/jobs --output-dir=$(repo_root)/prow/gcp/cluster/jobs write)
	@go run tools/prowtrans/cmd/prowtrans/main.go --requirement-presets=./prow/gcp/config/jobs/.base.yaml --configs=./prow/gcp/config/istio-private_jobs --input=./prow/gcp/config/jobs
	@go run tools/prowtrans/cmd/prowtrans/main.go --requirement-presets=./prow/gcp/config/jobs/.base.yaml --configs=./prow/gcp/config/experimental --input=./prow/gcp/config/jobs

# Mirror the canonical job config into the EKS (AWS) cluster tree consumed by the -aws deploy targets.
# EKS has no separate arm cluster: arm64 is a node group inside the default build cluster (prow-build),
# selected via nodeSelector. Strip the GKE-only `cluster: prow-arm` override so those jobs land there.
generate-config-aws: generate-config
	@rm -fr prow/aws/cluster/jobs/*/*/*.gen.yaml
	@(cd tools/prowgen/cmd/prowgen; go run main.go --input-dir=$(repo_root)/prow/aws/config/jobs --output-dir=$(repo_root)/prow/aws/cluster/jobs write)
	@go run tools/prowtrans/cmd/prowtrans/main.go --requirement-presets=./prow/aws/config/jobs/.base.yaml --configs=./prow/aws/config/istio-private_jobs --input=./prow/aws/config/jobs
	# experimental jobs are intentionally not generated for EKS (config removed)
	# @go run tools/prowtrans/cmd/prowtrans/main.go --requirement-presets=./prow/aws/config/jobs/.base.yaml --configs=./prow/aws/config/experimental --input=./prow/aws/config/jobs



diff-config:
	@(cd tools/prowgen/cmd/prowgen; GOARCH=$(GOARCH) GOOS=$(GOOS) go run main.go --input-dir=$(repo_root)/prow/gcp/config/jobs --output-dir=$(repo_root)/prow/cluster/gcp/jobs diff)

diff-config-aws:
	@(cd tools/prowgen/cmd/prowgen; GOARCH=$(GOARCH) GOOS=$(GOOS) go run main.go --input-dir=$(repo_root)/prow/aws/config/jobs --output-dir=$(repo_root)/prow/cluster/aws/jobs diff)

include common/Makefile.common.mk

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

gen: generate-config fmt mirror-licenses

gen-check: gen check-clean-repo

generate-config:
	@rm -fr prow/cluster/jobs/istio/*/*.gen.yaml
	@(cd tools/prowgen/cmd/prowgen; go run main.go --input-dir=$(repo_root)/prow/config/jobs --output-dir=$(repo_root)/prow/cluster/jobs write)
	@rm -fr prow/cluster/jobs/istio-private/*/*.gen.yaml
	@go run tools/prowtrans/cmd/prowtrans/main.go --configs=./prow/config/istio-private_jobs --input=./prow/config/jobs
	@go run tools/prowtrans/cmd/prowtrans/main.go --configs=./prow/config/experimental --input=./prow/config/jobs

diff-config:
	@(cd tools/prowgen/cmd/prowgen; GOARCH=$(GOARCH) GOOS=$(GOOS) go run main.go --input-dir=$(repo_root)/prow/config/jobs --output-dir=$(repo_root)/prow/cluster/jobs diff)

include common/Makefile.common.mk

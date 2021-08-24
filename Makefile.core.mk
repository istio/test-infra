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

lint: lint-all

lint-buildifier:
	@bazel run //:buildifier -- -showlog -mode=check $(git ls-files| grep -e BUILD -e WORKSPACE | grep -v vendor)

fmt: format-go tidy-go

test:
	@go test -race ./...

gen: generate-config fmt mirror-licenses

gen-check: gen check-clean-repo

generate-config:
	@rm -fr prow/cluster/jobs/istio/*/*.gen.yaml
	@(cd prow/config/cmd; go run generate.go write)
	@rm -fr prow/cluster/jobs/istio-private/*/*.gen.yaml
	@go run prow/genjobs/main.go --configs=./prow/config/istio-private_jobs
	@go run prow/genjobs/main.go --configs=./prow/config/istio-private_jobs

diff-config:
	@(cd prow/config/cmd; GOARCH=$(GOARCH) GOOS=$(GOOS) go run generate.go diff)

include common/Makefile.common.mk

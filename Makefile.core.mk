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

lint: lint-go

test:
	@go test -race ./...

.PHONY: testgrid
testgrid:
	configurator --prow-config prow/config.yaml --prow-job-config prow/cluster/jobs --output-yaml --yaml testgrid/default.yaml --oneshot --output testgrid/istio.gen.yaml

generate-config:
	@(cd prow/config/cmd; GOARCH=amd64 GOOS=linux go run generate.go write)

diff-config:
	@(cd prow/config/cmd; GOARCH=amd64 GOOS=linux go run generate.go diff)

check-config:
	@(cd prow/config/cmd; GOARCH=amd64 GOOS=linux go run generate.go check)

include common/Makefile.common.mk

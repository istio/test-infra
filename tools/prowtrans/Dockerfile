# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.23 as build-env

WORKDIR /go/src/istio

COPY . /go/src/istio

RUN CGO_ENABLED=0 go build -o /go/bin/prowtrans /go/src/istio/tools/prowtrans/cmd/prowtrans

# safer to use latest than never update...
# hadolint ignore=DL3006
FROM gcr.io/distroless/static

COPY --from=build-env /go/bin/prowtrans /

ENTRYPOINT ["/prowtrans"]

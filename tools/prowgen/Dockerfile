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

FROM golang:1.17 as build-env

WORKDIR /go/src/istio

COPY . /go/src/istio

RUN go install /go/src/istio/tools/prowgen/cmd/prowgen

RUN go install github.com/mikefarah/yq/v4@v4.16.2

FROM gcr.io/distroless/static:22bd467b41e5e656e31db347265fae118db166d9

COPY --from=build-env /go/bin/prowgen /bin

COPY --from=build-env /go/bin/yq /bin

ENV PATH=/bin

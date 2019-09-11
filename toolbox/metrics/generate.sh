#!/bin/bash

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

set -ex

bazel run //toolbox/metrics:metrics_fetcher

# shellcheck disable=SC2155
export GIT_HEAD=$(git rev-parse HEAD)

docker tag bazel/toolbox/metrics:metrics_fetcher gcr.io/istio-testing/metrics_fetcher:"${GIT_HEAD}"

docker push gcr.io/istio-testing/metrics_fetcher:"${GIT_HEAD}"

sed 's/image: gcr.io\/istio-testing\/metrics_fetcher:.*/image: gcr.io\/istio-testing\/metrics_fetcher:'"$GIT_HEAD"'/g' toolbox/metrics/metrics_fetcher.yaml > toolbox/metrics/local.metrics_fetcher.yaml

cat toolbox/metrics/local.metrics_fetcher.yaml

echo "To replace: kubectl replace -f toolbox/metrics/local.metrics_fetcher.yaml"

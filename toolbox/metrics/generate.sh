#!/bin/bash

set -ex

bazel run //toolbox/metrics:metrics_fetcher

export GIT_HEAD=$(git rev-parse HEAD)

docker tag bazel/toolbox/metrics:metrics_fetcher gcr.io/istio-testing/metrics_fetcher:${GIT_HEAD}

gcloud docker -- push gcr.io/istio-testing/metrics_fetcher:${GIT_HEAD}

sed 's/image: gcr.io\/istio-testing\/metrics_fetcher:.*/image: gcr.io\/istio-testing\/metrics_fetcher:'"$GIT_HEAD"'/g' toolbox/metrics/metrics_fetcher.yaml > toolbox/metrics/local.metrics_fetcher.yaml

cat toolbox/metrics/local.metrics_fetcher.yaml

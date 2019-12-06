#!/usr/bin/env bash

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

set -o errexit
set -o nounset
set -o pipefail

COMMON_OPTS=(
  "--mapping=istio=istio-private"
  "--ssh-clone"
  "--extra-refs"
  "--input=./cluster/jobs/"
  "--output=./cluster/jobs/"
  "--bucket=istio-private-build"
  "--ssh-key-secret=ssh-key-secret"
  "--cluster=private"
  "--modifier=priv"
)

# Clean ./prow/cluster/jobs/istio-private directory
go run ./genjobs --clean --mapping=istio=istio-private --output=./cluster/jobs/ --dry-run >/dev/null

# istio/istio build job(s) - postsubmit(s)
go run ./genjobs \
  "${COMMON_OPTS[@]}" \
  --branches=release-1.4,master \
  --env DOCKER_HUB=gcr.io/istio-prow-build,GCS_BUCKET=istio-private-build/dev \
  --labels preset-enable-ssh=true \
  --job-type postsubmit \
  --repo-whitelist istio \
  --job-whitelist release_istio_postsubmit,release_istio_release-1.4_postsubmit

# istio/istio test jobs(s) - presubmit(s) and postsubmit(s)
go run ./genjobs \
  "${COMMON_OPTS[@]}" \
  --branches=release-1.4,master \
  --labels preset-override-envoy=true \
  --job-type presubmit,postsubmit \
  --repo-whitelist istio \
  --job-blacklist release_istio_postsubmit,release_istio_release-1.4_postsubmit

# istio/proxy master test jobs(s) - presubmit(s)
go run ./genjobs \
  "${COMMON_OPTS[@]}" \
  --branches master \
  --modifier=master_priv \
  --labels preset-enable-netrc=true \
  --job-type presubmit \
  --env BAZEL_BUILD_RBE_JOBS=0,ENVOY_REPOSITORY=https://github.com/envoyproxy/envoy-wasm,ENVOY_PREFIX=envoy-wasm \
  --repo-whitelist proxy

# istio/proxy master build jobs(s) - postsubmit(s)
go run ./genjobs \
  "${COMMON_OPTS[@]}" \
  --branches master \
  --modifier master_priv \
  --labels preset-enable-netrc=true \
  --job-type postsubmit \
  --env BAZEL_BUILD_RBE_JOBS=0,GCS_BUILD_BUCKET=istio-private-build,GCS_ARTIFACTS_BUCKET=istio-private-artifacts,DOCKER_REPOSITORY=istio-prow-build/envoy,ENVOY_REPOSITORY=https://github.com/envoyproxy/envoy-wasm,ENVOY_PREFIX=envoy-wasm \
  --repo-whitelist proxy

# istio/proxy release-1.4 test jobs(s) - presubmit(s)
go run ./genjobs \
  "${COMMON_OPTS[@]}" \
  --branches release-1.4 \
  --modifier release-1.4_priv \
  --labels preset-enable-netrc=true \
  --job-type presubmit \
  --env BAZEL_BUILD_RBE_JOBS=0,ENVOY_REPOSITORY=https://github.com/istio-private/envoy,ENVOY_PREFIX=envoy \
  --repo-whitelist proxy

# istio/proxy release-1.4 build jobs(s) - postsubmit(s)
go run ./genjobs \
  "${COMMON_OPTS[@]}" \
  --branches release-1.4 \
  --modifier=release-1.4_priv \
  --labels preset-enable-netrc=true \
  --job-type postsubmit \
  --env BAZEL_BUILD_RBE_JOBS=0,GCS_BUILD_BUCKET=istio-private-build,GCS_ARTIFACTS_BUCKET=istio-private-artifacts,DOCKER_REPOSITORY=istio-prow-build/envoy,ENVOY_REPOSITORY=https://github.com/istio-private/envoy,ENVOY_PREFIX=envoy \
  --repo-whitelist proxy

# istio/release-builder master test jobs(s) - pre/postsubmit(s)
go run ./genjobs \
  "${COMMON_OPTS[@]}" \
  --branches=master \
  --job-type presubmit,postsubmit \
  --repo-whitelist release-builder \
  --job-whitelist lint_release-builder,lint_release-builder_postsubmit,test_release-builder,test_release-builder_postsubmit,gencheck_release-builder,gencheck_release-builder_postsubmit

# istio/release-builder release-1.4 test jobs(s) - pre/postsubmit(s)
go run ./genjobs \
  "${COMMON_OPTS[@]}" \
  --branches=release-1.4 \
  --job-type presubmit,postsubmit \
  --repo-whitelist release-builder \
  --job-whitelist lint_release-builder_release-1.4,lint_release-builder_release-1.4_postsubmit,test_release-builder_release-1.4,test_release-builder_release-1.4_postsubmit,gencheck_release-builder_release-1.4,gencheck_release-builder_release-1.4_postsubmit

# istio/release-builder build warning jobs(s) - presubmit(s)
go run ./genjobs \
  "${COMMON_OPTS[@]}" \
  --branches=release-1.4,master \
  --env PRERELEASE_DOCKER_HUB=gcr.io/istio-prow-build,GCS_BUCKET=istio-private-prerelease/prerelease \
  --job-type presubmit \
  --repo-whitelist release-builder \
  --job-whitelist build-warning_release-builder,build-warning_release-builder_release-1.4

# istio/release-builder build jobs(s) - postsubmit(s)
go run ./genjobs \
  "${COMMON_OPTS[@]}" \
  --branches=release-1.4,master \
  --labels preset-enable-ssh=true,preset-override-envoy=true,preset-override-deps=release-1.4 \
  --env PRERELEASE_DOCKER_HUB=gcr.io/istio-prow-build,GCS_BUCKET=istio-private-prerelease/prerelease \
  --job-type postsubmit \
  --repo-whitelist release-builder \
  --job-whitelist build-release_release-builder_release-1.4_postsubmit,build-release_release-builder_postsubmit

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

set -eux

export CLOUDSDK_PYTHON_SITEPACKAGES=1
export CLOUDSDK_CORE_DISABLE_PROMPTS=1
export CLOUDSDK_INSTALL_DIR=/usr/lib/
curl https://sdk.cloud.google.com | bash

export PATH="${PATH}:/usr/lib/google-cloud-sdk/bin"

sed -i -e 's/true/false/' /usr/lib/google-cloud-sdk/lib/googlecloudsdk/core/config.json
gcloud -q components update


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

cd "$(dirname "$0")" || exit

if [[ ! -f install.sh ]]; then
  echo "ERROR: install.sh not found" >&1
  echo "  curl https://raw.githubusercontent.com/kubernetes/test-infra/master/rbe/install.sh -o install.sh" >&1
  exit 1
fi

if [[ ! -f configure.sh ]]; then
  echo "ERROR: configure.sh not found" >&1
  echo "  curl https://raw.githubusercontent.com/kubernetes/test-infra/master/rbe/configure.sh -o configure.sh" >&1
  exit 1
fi


proj=istio-testing
pool=prow-pool
workers=200
disk=200
machine=n1-standard-2
bot=istio-prow-test-job@istio-testing.iam.gserviceaccount.com

./install.sh "$proj" "$pool" "$workers" "$disk" "$machine" "$bot"

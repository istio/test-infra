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

echo "# Run update-prow-staging.sh to keep sync this with the master config" > istio.istio.prow-staging.yaml
cat istio.istio.master.yaml >> istio.istio.prow-staging.yaml
sed -i 's/-master/-prow-staging/g' istio.istio.prow-staging.yaml
sed -i 's/\^master/\^prow-staging/g' istio.istio.prow-staging.yaml
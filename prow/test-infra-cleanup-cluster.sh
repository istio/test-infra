#!/bin/bash

# Copyright 2017 Istio Authors

#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at

#       http://www.apache.org/licenses/LICENSE-2.0

#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.


#############################################
# Crontab cleanup script triggered by Prow. #
#############################################

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

if [ "${CI:-}" == 'bootstrap' ]; then
    # Test harness will checkout code to directoryc$GOPATH/src/github.com/istio
    # but we depend on being at path $GOPATH/src/istio.io for imports.
    mkdir -p ${GOPATH}/src/istio.io
    ln -s ${GOPATH}/src/github.com/istio/test-infra ${GOPATH}/src/istio.io
    cd ${GOPATH}/src/istio.io/test-infra/
fi

echo "=== Cleaning up cluster ==="
./scripts/cleanup-cluster -h 2

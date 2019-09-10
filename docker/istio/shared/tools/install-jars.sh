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

JARS_DIR="/usr/local/bin/flakey_jars"
wget -q -nc -O jars.zip https://github.com/ChristinaLyu0710/istio-flakey-test/raw/master/FlakeyTest/flexible/Flakey/jars.zip
mkdir -p "$JARS_DIR"
unzip jars.zip -d "$JARS_DIR"

export PATH=${PATH}:/usr/local/bin
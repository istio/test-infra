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

# shellcheck disable=SC2046
# shellcheck disable=SC2154
gcloud auth application-default login; g=0; for n in $(gsutil ls {gs://istio-circleci/master/*/*/artifacts/junit.xml,gs://istio-prow/logs/*master/*/artifacts/junit.xml}); do foo=$(cut -d "," -f 2 <<< $(cut -d ":" -f 2 <<< $(gsutil stat "$n" | sed -n 3p))); gsutil cp "$n" "gs://istio-flakey-test/$data_folder/out-$foo-$g-master.xml"; ((++g)); done; g=0; for n in $(gsutil ls {gs://istio-circleci/release-1.2/*/*/artifacts/junit.xml,gs://istio-prow/logs/*release-1.2/*/artifacts/junit.xml}); do foo=$(cut -d "," -f 2 <<< $(cut -d ":" -f 2 <<< $(gsutil stat "$n" | sed -n 3p))); gsutil cp "$n" "gs://istio-flakey-test/$data_folder/out-$foo-$g-release1.2.xml"; ((++g)); done
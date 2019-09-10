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

g=0; for n in $(gsutil ls gs://istio-circleci/master/test-integration-kubernetes/*/artifacts/junit.xml); do foo=$(cut -d "," -f 2 <<< $(cut -d ":" -f 2 <<< $(gsutil stat $n | sed -n 3p))); echo $n; gsutil cp $n "gs://istio-flakey-test/temp/out-$foo-$g-master.xml"; ((++g)); done; g=0; for n in $(gsutil ls gs://istio-circleci/release-1.2/e2e-mixer-noauth-v1alpha3-v2/*/artifacts/junit.xml); do foo=$(cut -d "," -f 2 <<< $(cut -d ":" -f 2 <<< $(gsutil stat $n | sed -n 3p))); echo $n; gsutil cp $n "gs://istio-flakey-test/temp/out-$foo-$g-release1.2.xml"; ((++g)); done


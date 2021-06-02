#!/usr/bin/env bash

# Copyright 2020 Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ROOT="$(cd -P "$(dirname -- "$0")" && pwd -P)"

# shellcheck disable=SC1090
source "$ROOT/lib.sh"

timout="5m"

get_tokeninfo() {
  local token

  while [ -z "${token:-}" ]; do
    sleep 5
    token=$(kubectl get secrets authentikos-token --output=jsonpath='{.data.token}' | base64 --decode)
  done

  curl -sSfL "https://www.googleapis.com/oauth2/v1/tokeninfo?access_token=$token"
}

run_test() { (
  set -ex

  local tokeninfo="$1"

  echo "Test 'tokeninfo' response body exists"
  test -n "$tokeninfo"

  echo "Test 'error' is null"
  test "$(echo "$tokeninfo" | jq -r '.error')" = "null"

  echo "Test 'issued_to' exists"
  test -n "$(echo "$tokeninfo" | jq -r '.issued_to')"

  echo "Test 'audience' exists"
  test -n "$(echo "$tokeninfo" | jq -r '.audience')"

  echo "Test 'user_id' exists"
  test -n "$(echo "$tokeninfo" | jq -r '.user_id')"

  echo "Test 'email' exists"
  test -n "$(echo "$tokeninfo" | jq -r '.email')"

  echo "Test 'expires_in' greater than 0"
  test "$(echo "$tokeninfo" | jq -r '.expires_in')" -gt 0

  echo "Test 'scope' includes 'userinfo.email', 'cloud-platform', and 'openid'"
  echo "$tokeninfo" | jq -r '.scope' |
    grep -w "openid" |
    grep -w "https://www.googleapis.com/auth/cloud-platform" |
    grep -w "https://www.googleapis.com/auth/userinfo.email" >/dev/null
); }

main() {

  echo "Try curling https://accounts.google.com/o/oauth2/token outside kind"
  curl -v  https://accounts.google.com/o/oauth2/token || echo "FAILED"

  echo "Try curling https://accounts.google.com/o/oauth2/token from kind"
  kubectl run -i --rm hello --image=curlimages/curl --restart=Never -- curl -v https://accounts.google.com/o/oauth2/token || echo "FAILED"


  echo "Beginning setup"
  kubectl create secret generic service-account --from-file="service-account.json=$GOOGLE_APPLICATION_CREDENTIALS"
  kubectl apply --filename="$ROOT/testdata/authentikos-simple.yaml"
  kubectl wait deployment authentikos --for="condition=available" --timeout="$timout"

  echo "Begin tests"
  # Unset "errexit" to allow execution to continue if "run_test" fails.
  set +e
  run_test "$(with_timeout get_tokeninfo "$timout")"; exit_code="$?"
  set -e

  echo "Dumping authentikos logs VV"
  kubectl logs -l app=authentikos
}

main
exit "$exit_code"
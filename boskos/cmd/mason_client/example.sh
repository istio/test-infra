#!/bin/bash

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

FILE_LOG=file.log

function cleanup() {
  kill -SIGINT ${PID}
  wait
}

function wait_10mn() {
  for i in `seq 1 60`; do
    grep -q READY $FILE_LOG && return 0
    kill -s 0 ${PID} || return 1
    sleep 10
  done
  return 1
}

go install istio.io/test-infra/boskos/cmd/mason_client
mason_client \
  --type gke-e2e-test \
  --boskos-url http://35.202.113.249 \
  --owner sebvas \
  --info-save info.save \
  --kubeconfig-save kubeconfig.save > ${FILE_LOG} 2>&1 &

PID=$!

trap cleanup EXIT

wait_10mn

echo "start e2e test here"


#!/bin/bash

cd "$(dirname "$0")"

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

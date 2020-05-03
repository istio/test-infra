#!/usr/bin/env bash
# (re) create job-config of Prow using yaml files under test-infra/prow/cluster/jobs/
#
kubectl create configmap config --from-file=config.yaml=./config.yaml --dry-run=client -o yaml | kubectl replace configmap config -f -
kubectl create configmap plugins --from-file=plugins.yaml=./plugins.yaml --dry-run=client -o yaml |kubectl replace configmap plugins -f -


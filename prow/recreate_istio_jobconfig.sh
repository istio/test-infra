#!/usr/bin/env bash
# (re) create job-config of Prow using yaml files under test-infra/prow/cluster/jobs/
#
kubectl create configmap job-config --from-file=all-presets.yaml=cluster/jobs/all-presets.yaml \
 --from-file=istio.istio.release-1.6.gen.yaml=cluster/jobs/istio/istio/istio.istio.release-1.6.gen.yaml \
 --from-file=istio.istio.master.gen.yaml=cluster/jobs/istio/istio/istio.istio.master.gen.yaml \
 --from-file=istio.istio.release-1.4.gen.yaml=cluster/jobs/istio/istio/istio.istio.release-1.4.gen.yaml \
 --from-file=istio.istio.release-1.5.gen.yaml=cluster/jobs/istio/istio/istio.istio.release-1.5.gen.yaml \
 --dry-run=client -o yaml | kubectl replace configmap job-config -f -

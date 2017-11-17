#!/bin/bash

set -ex

bazel run //toolbox/oncall_alert:oncall_alert_image

export GIT_HEAD=$(git rev-parse HEAD)

docker tag bazel/toolbox/oncall_alert:oncall_alert_image gcr.io/istio-testing/oncall_alert:${GIT_HEAD}

gcloud docker -- push gcr.io/istio-testing/oncall_alert:${GIT_HEAD}

sed 's/image: gcr.io\/istio-testing\/oncall_alert:.*/image: gcr.io\/istio-testing\/oncall_alert:'"$GIT_HEAD"'/g' toolbox/oncall_alert/alert-deployment.yaml > toolbox/oncall_alert/local.alert-deployment.yaml

cat toolbox/oncall_alert/local.alert-deployment.yaml

echo "To replace: 1. connect to \"istio-toolbox\" cluster and 2. kubectl replace -f toolbox/oncall_alert/local.alert-deployment.yaml -n oncall-alert"

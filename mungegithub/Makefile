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

PROJECT ?= istio-testing
ZONE    ?= us-west1-a
CLUSTER ?= istio-mungegithub
REPO    ?= test-infra

# ensures that kubectl has access to the prow cluster's config context
get-cluster-credentials:
	gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT)" --zone="$(ZONE)"


deploy: get-cluster-credentials
	sed -e 's!@@!$(REPO)!g' ./deployment/deployment.yaml > ./deployment/$(REPO)/local.deployment.yaml
	kubectl apply -f ./deployment/$(REPO)/local.deployment.yaml


update-config: get-cluster-credentials
	make stopd
	kubectl create configmap $(REPO)-sq-config --from-file=config=./deployment/$(REPO)/configmap.yaml --dry-run -o yaml | kubectl replace configmap config -f -
	make startd

startd:
	REPO=$(REPO) make deploy


stopd:
	kubectl delete deployment $(REPO)-submit-queue

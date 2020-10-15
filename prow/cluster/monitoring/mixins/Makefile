# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# The targets defined here are only for local testing and debuging
# There is no intention of using it in CI/CD pipeline for the moment

SHELL=/bin/bash -o pipefail

include ../../../Makefile.gcloud.mk

PROJECT ?= istio-testing
CLUSTER ?= prow
ZONE    ?= us-west1-a

grafana-dashboards:
	@mkdir -p dashboards_out
	@for input in $(shell ls grafana_dashboards); do \
		output="$${input%.*}.json"; \
		echo "Generating $${output} from $${input} ..."; \
		jsonnet -J vendor/ "grafana_dashboards/$${input}" > "./dashboards_out/$${output}"; \
	done

prow_prometheusrule.yaml:
	@mkdir -p prometheus_out
	@echo "Generating prow_prometheusrule.yaml ..."
	jsonnet ./prometheus/prow_prometheusrule.jsonnet | yq r -P - > prometheus_out/$@

apply-configmaps: get-cluster-credentials grafana-dashboards
	@for input in $(shell ls dashboards_out); do \
		dashboard_name="grafana-dashboard-$${input%.*}"; \
		echo "Generating dashboard $${dashboard_name} from $${input} ..."; \
		kubectl create configmap -n prow-monitoring "$${dashboard_name}" --from-file="$${input}=dashboards_out/$${input}" --dry-run -o yaml | kubectl apply -f -; \
	done

apply-prow-prometheusrule: get-cluster-credentials prow_prometheusrule.yaml
	kubectl apply -f prometheus_out/

apply: apply-configmaps apply-prow-prometheusrule

clean:
	rm -rfv ./vendor
	rm -rfv ./dashboards_out/*.json
	rm -rfv ./prometheus_out/*.json
	rm -rfv ./prometheus_out/*.yaml

install:
	jb install
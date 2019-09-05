# Copyright Istio Authors 2017.
#
# Based off a similar file used to manage the k8s cluster
# https://github.com/kubernetes/test-infra/blob/master/prow/Makefile

# GKE variables.
PROJECT ?= istio-testing
PROJECT_BUILD ?= istio-prow-build
ZONE    ?= us-west1-a
CLUSTER ?= prow

# ensures that kubectl has access to the prow cluster's config context
get-cluster-credentials:
	# TODO(fejta): https://github.com/istio/test-infra/issues/1636
	gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT)" --zone="$(ZONE)"

get-build-cluster-credentials:
	gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT_BUILD)" --zone="$(ZONE)"

update-config-dry-run: get-cluster-credentials
	./recreate_prow_configmaps.py \
		--job-config-dir=cluster/jobs \
		--prow-config-path=config.yaml \
		--plugins-config-path=plugins.yaml

update-config: get-cluster-credentials
	./recreate_prow_configmaps.py \
		--job-config-dir=cluster/jobs \
		--prow-config-path=config.yaml \
		--plugins-config-path=plugins.yaml \
		--wet

deploy: get-cluster-credentials
	kubectl "--context=gke_$(PROJECT)_$(ZONE)_$(CLUSTER)" apply -f ./cluster/

deploy-build: get-build-cluster-credentials
	kubectl "--context=gke_$(PROJECT_BUILD)_$(ZONE)_$(CLUSTER)" apply -f ./cluster/build/

.PHONY: deploy deploy-build update-config update-config-dry-run get-cluster-credentials get-build-cluster-credentials

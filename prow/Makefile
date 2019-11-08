# Copyright Istio Authors 2017.
#
# Based off a similar file used to manage the k8s cluster
# https://github.com/kubernetes/test-infra/blob/master/prow/Makefile

include Makefile.gcloud.mk

# GKE variables.
PROJECT ?= istio-testing
PROJECT_BUILD ?= istio-prow-build
ZONE    ?= us-west1-a
CLUSTER ?= prow

generate-istio-private-jobs:
	go run ./genjobs \
		--clean \
		--mapping istio=istio-private \
		--job-type postsubmit \
		--channel - \
		--env DOCKER_HUB=gcr.io/istio-prow-build,GCS_BUCKET=istio-private-build/dev \
		--input ./cluster/jobs/ \
		--output ./cluster/jobs/ \
		--repo-whitelist istio \
		--job-whitelist release_istio_postsubmit \
		--bucket istio-private-build \
		--cluster private \
		--selector testing=build-pool

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
	kubectl apply -f ./cluster/

deploy-build: get-build-cluster-credentials
	kubectl apply -f ./cluster/build/

.PHONY: generate-istio-private-jobs deploy deploy-build update-config update-config-dry-run

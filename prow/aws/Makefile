# Copyright Istio Authors 2017.
#
# Based off a similar file used to manage the k8s cluster
# https://github.com/kubernetes/test-infra/blob/master/prow/Makefile

include Makefile.aws.mk

# AWS variables.
REGION ?= us-west-2

CLUSTER ?= prow
CLUSTER_AWS ?= prow
CLUSTER_BUILD_AWS ?= prow-build
CLUSTER_PRIVATE_AWS ?= prow-private

create-istio-deps-configmap: branch ?= master
create-istio-deps-configmap: get-private-cluster-credentials
	@bash ./create-deps-cm.sh \
	--local \
	--branch="$(branch)" \
	--namespace=test-pods \
	--key=dependencies \
	$(if $(filter %,$(dry_run)),--dry-run,) \
	"$(branch)-istio-deps"

create-release-deps-configmap: branch ?= master
create-release-deps-configmap: get-private-cluster-credentials
	@bash ./create-deps-cm.sh \
	--branch="$(branch)" \
	--namespace=test-pods \
	--key=dependencies \
	$(if $(filter %,$(dry_run)),--dry-run,) \
	"$(branch)-release-deps"

create-deps-configmaps: create-istio-deps-configmap create-release-deps-configmap

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
# Apply the ProwJob CRD with --server-side=true due to its size.
	kubectl apply --server-side=true -f ./cluster/prowjob-crd/
	kubectl apply -f ./cluster/

deploy-gcsweb: get-cluster-credentials
	kubectl apply -f ./cluster/gcsweb/ -n gcs

get-build-cluster-credentials: CLUSTER=$(CLUSTER_BUILD_AWS)

deploy-build: get-build-cluster-credentials
	kubectl apply -f ./cluster/build/

query-build: get-build-cluster-credentials
	kubectl cluster-info

get-private-cluster-credentials: CLUSTER=$(CLUSTER_PRIVATE_AWS)

deploy-private: get-private-cluster-credentials
	kubectl apply -f ./cluster/private/

query-private: get-private-cluster-credentials
	kubectl cluster-info

.PHONY: deploy deploy-gcsweb deploy-build deploy-private query-build query-private update-config update-config-dry-run

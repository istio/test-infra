# Copyright Istio Authors 2017.
#
# Based off a similar file used to manage the k8s cluster
# https://github.com/kubernetes/test-infra/blob/master/prow/Makefile

include Makefile.gcloud.mk

# GKE variables.
PROJECT ?= istio-testing
PROJECT_ARM ?= istio-prow-build
PROJECT_BUILD ?= istio-prow-build
PROJECT_PRIVATE ?= istio-prow-build
CLUSTER ?= prow
CLUSTER_ARM ?= prow-arm
CLUSTER_BUILD ?= prow
CLUSTER_PRIVATE ?= prow-private
ZONE    ?= us-west1-a
ZONE_ARM ?= us-central1-f
ZONE_BUILD ?= us-west1-a
ZONE_PRIVATE ?= us-west1-a

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

# TODO(clarketm): enable pruning
deploy-monitoring: get-cluster-credentials
	kubectl apply -f ./cluster/monitoring/prometheus-operator-crds/
	$(MAKE) -C cluster/monitoring/mixins clean apply
	kubectl apply -f ./cluster/monitoring/

get-build-cluster-credentials: PROJECT=$(PROJECT_BUILD)
get-build-cluster-credentials: CLUSTER=$(CLUSTER_BUILD)
get-build-cluster-credentials: ZONE=$(ZONE_BUILD)

deploy-build: get-build-cluster-credentials
	kubectl apply -f ./cluster/build/

query-build: get-build-cluster-credentials
	kubectl cluster-info

get-arm-cluster-credentials: PROJECT=$(PROJECT_ARM)
get-arm-cluster-credentials: CLUSTER=$(CLUSTER_ARM)
get-arm-cluster-credentials: ZONE=$(ZONE_ARM)

deploy-arm: get-arm-cluster-credentials
	kubectl apply -f ./cluster/arm

query-arm: get-arm-cluster-credentials
	kubectl cluster-info

get-private-cluster-credentials: PROJECT=$(PROJECT_PRIVATE)
get-private-cluster-credentials: CLUSTER=$(CLUSTER_PRIVATE)
get-private-cluster-credentials: ZONE=$(ZONE_PRIVATE)

deploy-private: get-private-cluster-credentials
	kubectl apply -f ./cluster/private/

query-private: get-private-cluster-credentials
	kubectl cluster-info

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

.PHONY: gen-private-jobs deploy deploy-gcsweb deploy-build deploy-private update-config update-config-dry-run create-istio-deps-configmap create-release-deps-configmap create-deps-configmaps

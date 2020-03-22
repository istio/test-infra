# Copyright Istio Authors 2017.
#
# Based off a similar file used to manage the k8s cluster
# https://github.com/kubernetes/test-infra/blob/master/prow/Makefile

include Makefile.gcloud.mk

# GKE variables.
PROJECT ?= istio-testing
PROJECT_BUILD ?= istio-prow-build
PROJECT_PRIVATE ?= istio-prow-build
CLUSTER ?= prow
CLUSTER_BUILD ?= prow
CLUSTER_PRIVATE ?= prow-private
ZONE    ?= us-west1-a
export PRUNE_WL

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

deploy: get-cluster-credentials get-api-resources
	kubectl apply -f ./cluster/ --prune -l app.kubernetes.io/part-of=prow $(PRUNE_WL)

deploy-gcsweb: get-cluster-credentials get-api-resources
	kubectl apply -f ./cluster/gcsweb/ --prune -l app.kubernetes.io/part-of=gcsweb $(PRUNE_WL)

deploy-velodrome: get-cluster-credentials get-api-resources
	kubectl apply -f ./cluster/velodrome/ --prune -l app.kubernetes.io/part-of=velodrome $(PRUNE_WL)

get-build-cluster-credentials: PROJECT=$(PROJECT_BUILD)
get-build-cluster-credentials: CLUSTER=$(CLUSTER_BUILD)

deploy-build: get-build-cluster-credentials get-build-api-resources
	kubectl apply -f ./cluster/build/ --prune -l app.kubernetes.io/part-of=prow-build $(PRUNE_WL)

get-private-cluster-credentials: PROJECT=$(PROJECT_PRIVATE)
get-private-cluster-credentials: CLUSTER=$(CLUSTER_PRIVATE)

deploy-private: get-private-cluster-credentials get-private-api-resources
	kubectl apply -f ./cluster/private/ --prune -l app.kubernetes.io/part-of=prow-private $(PRUNE_WL)

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

.PHONY: gen-private-jobs deploy deploy-gcsweb deploy-velodrome deploy-build deploy-private update-config update-config-dry-run create-istio-deps-configmap create-release-deps-configmap create-deps-configmaps

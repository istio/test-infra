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
	kubectl apply -f ./cluster/ --prune -l app.kubernetes.io/part-of=prow

deploy-gcsweb: get-cluster-credentials
	kubectl apply -f ./cluster/gcsweb/ --prune -l app.kubernetes.io/part-of=gcsweb

deploy-velodrome: get-cluster-credentials
	kubectl apply -f ./cluster/velodrome/ --prune -l app.kubernetes.io/part-of=velodrome

deploy-build: PROJECT=$(PROJECT_BUILD)
deploy-build: CLUSTER=$(CLUSTER_BUILD)
deploy-build: get-build-cluster-credentials
	kubectl apply -f ./cluster/build/ --prune -l app.kubernetes.io/part-of=prow-build

deploy-private: PROJECT=$(PROJECT_PRIVATE)
deploy-private: CLUSTER=$(CLUSTER_PRIVATE)
deploy-private: get-private-cluster-credentials
	kubectl apply -f ./cluster/private/ --prune -l app.kubernetes.io/part-of=prow-private

.PHONY: gen-private-jobs deploy deploy-gcsweb deploy-velodrome deploy-build deploy-private update-config update-config-dry-run

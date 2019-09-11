# Copyright Istio Authors 2017.
#
# Based off a similar file used to manage the k8s cluster
# https://github.com/kubernetes/test-infra/blob/master/prow/Makefile

# GKE variables.
PROJECT ?= istio-testing
PROJECT_BUILD ?= istio-prow-build
ZONE    ?= us-west1-a
CLUSTER ?= prow

export KUBECONFIG

# https://github.com/istio/test-infra/issues/1636
# This prevents the Kube current-context in the execution environment from being
# overwritten unless the intention is made explicit w/ the `save` parameter.
# e.g.
#		make get-cluster-credentials save=true
save-kubeconfig:
ifndef save
	$(eval KUBECONFIG=$(shell mktemp))
endif

activate-serviceaccount:
ifdef GOOGLE_APPLICATION_CREDENTIALS
	gcloud auth activate-service-account --key-file="$(GOOGLE_APPLICATION_CREDENTIALS)"
endif

get-cluster-credentials: save-kubeconfig activate-serviceaccount
	gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT)" --zone="$(ZONE)"

get-build-cluster-credentials: save-kubeconfig activate-serviceaccount
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
	kubectl apply -f ./cluster/

deploy-build: get-build-cluster-credentials
	kubectl apply -f ./cluster/build/

.PHONY: deploy deploy-build update-config update-config-dry-run get-cluster-credentials get-build-cluster-credentials save-kubeconfig activate-serviceaccount

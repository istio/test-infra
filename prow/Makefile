# Copyright Istio Authors 2017.
#
# Based off a similar file used to manage the k8s cluster
# https://github.com/kubernetes/test-infra/blob/master/prow/Makefile

# GKE variables.
PROJECT ?= istio-testing
ZONE    ?= us-west1-a
CLUSTER ?= prow

# ensures that kubectl has access to the prow cluster's config context
get-cluster-credentials:
	gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT)" --zone="$(ZONE)"

update-config: get-cluster-credentials
	kubectl create configmap config --from-file=config=./config.yaml --dry-run -o yaml | kubectl replace configmap config -f -

update-plugins: get-cluster-credentials
	kubectl create configmap plugins --from-file=plugins=./plugins.yaml --dry-run -o yaml | kubectl replace configmap plugins -f -

# use to guarantee that on command completion all deployments see new configs
update-configs:
	make stopd
	make update-config
	make update-plugins
	make startd

.PHONY: update-config update-plugins get-cluster-credentials update-configs

deck-deployment: get-cluster-credentials
	kubectl apply -f cluster/deck_deployment.yaml

deck-service: get-cluster-credentials
	kubectl apply -f cluster/deck_service.yaml

hook-deployment: get-cluster-credentials
	kubectl apply -f cluster/hook_deployment.yaml

hook-service: get-cluster-credentials
	kubectl apply -f cluster/hook_service.yaml

sinker-deployment: get-cluster-credentials
	kubectl apply -f cluster/sinker_deployment.yaml

plank-deployment: get-cluster-credentials
	kubectl apply -f cluster/plank_deployment.yaml

tot-deployment: get-cluster-credentials
	kubectl apply -f cluster/tot_deployment.yaml

tot-service: get-cluster-credentials
	kubectl apply -f cluster/tot_service.yaml

.PHONY: hook-deployment hook-service sinker-deployment deck-deployment deck-service tot-service tot-deployment plank-deployment

# Deployment Management
#
# Assume that configuration is all set up and services are created

startd:
	make hook-deployment
	make deck-deployment
	make plank-deployment
	make tot-deployment
	make sinker-deployment

stopd:
	kubectl delete deployment hook deck plank tot sinker

restartd:
	make stopd
	make startd

.PHONY: startd stopd restartd

# ---

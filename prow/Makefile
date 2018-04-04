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
update-configs: update-config update-plugins

.PHONY: update-config update-plugins get-cluster-credentials update-configs

branch-protector-cronjob:
	kubectl apply -f cluster/branchprotector_cronjob.yaml

deck-deployment: get-cluster-credentials
	kubectl apply -f cluster/deck_deployment.yaml

hook-deployment: get-cluster-credentials
	kubectl apply -f cluster/hook_deployment.yaml

horologium-deployment: get-cluster-credentials
	kubectl apply -f cluster/horologium_deployment.yaml

needs-rebase-deployment: get-cluster-credentials
	kubectl apply -f cluster/needs-rebase_deployment.yaml

sinker-deployment: get-cluster-credentials
	kubectl apply -f cluster/sinker_deployment.yaml

plank-deployment: get-cluster-credentials
	kubectl apply -f cluster/plank_deployment.yaml

tide-deployment: get-cluster-credentials
	kubectl apply -f cluster/tide_deployment.yaml

tot-deployment: get-cluster-credentials
	kubectl apply -f cluster/tot_deployment.yaml

deployments: branch-protector-cronjob deck-deployment hook-deployment horologium-deployment needs-rebase-deployment plank-deployment sinker-deployment tide-deployment tot-deployment

.PHONY: branch-protector-cronjob deployments deck-deployment hook-deployment horologium-deployment needs-rebase-deployment plank-deployment sinker-deployment tide-deployment tot-deployment

deck-service: get-cluster-credentials
	kubectl apply -f cluster/deck_service.yaml

hook-service: get-cluster-credentials
	kubectl apply -f cluster/hook_service.yaml

needs-rebase-service: get-cluster-credentials
	kubectl apply -f cluster/needs-rebase_service.yaml

tide-service: get-cluster-credentials
	kubectl apply -f cluster/tide_service.yaml

tot-service: get-cluster-credentials
	kubectl apply -f cluster/tot_service.yaml

services: deck-service hook-service needs-rebase-service tide-service tot-service

.PHONY: services deck-service hook-service needs-rebase-service tide-service tot-service

deploy: deployments services

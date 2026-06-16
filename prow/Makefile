# Copyright Istio Authors 2017.
#
# Based off a similar file used to manage the k8s cluster
# https://github.com/kubernetes/test-infra/blob/master/prow/Makefile

include Makefile.gcloud.mk
include Makefile.aws.mk

# GKE variables.
PROJECT ?= istio-testing
PROJECT_ARM ?= istio-prow-build
PROJECT_BUILD ?= istio-prow-build
PROJECT_PRIVATE ?= istio-prow-private

CLUSTER ?= prow
CLUSTER_ARM ?= prow-arm
CLUSTER_BUILD ?= prow
CLUSTER_PRIVATE ?= prow

ZONE    ?= us-west1-a
ZONE_ARM ?= us-central1-f
ZONE_BUILD ?= us-west1-a
ZONE_PRIVATE ?= us-central1-f

# AWS variables.
REGION ?= us-west-2

CLUSTER_AWS ?= prow
CLUSTER_BUILD_AWS ?= prow-build
CLUSTER_PRIVATE_AWS ?= prow-private

update-config-dry-run: get-cluster-credentials
	./recreate_prow_configmaps.py \
		--job-config-dir=cluster/gke/jobs \
		--prow-config-path=config.yaml \
		--plugins-config-path=plugins.yaml

update-config: get-cluster-credentials
	./recreate_prow_configmaps.py \
		--job-config-dir=cluster/gke/jobs \
		--prow-config-path=config.yaml \
		--plugins-config-path=plugins.yaml \
		--wet

deploy: get-cluster-credentials
# Apply the ProwJob CRD with --server-side=true due to its size.
	kubectl apply --server-side=true -f ./cluster/gke/prowjob-crd/
	kubectl apply -f ./cluster/gke/

deploy-gcsweb: get-cluster-credentials
	kubectl apply -f ./cluster/gke/gcsweb/ -n gcs

# TODO(clarketm): enable pruning
deploy-monitoring: get-cluster-credentials
	kubectl apply -f ./cluster/gke/monitoring/prometheus-operator-crds/
	$(MAKE) -C cluster/gke/monitoring/mixins clean apply
	kubectl apply -f ./cluster/gke/monitoring/

get-build-cluster-credentials: PROJECT=$(PROJECT_BUILD)
get-build-cluster-credentials: CLUSTER=$(CLUSTER_BUILD)
get-build-cluster-credentials: ZONE=$(ZONE_BUILD)

deploy-build: get-build-cluster-credentials
	kubectl apply -f ./cluster/gke/build/

query-build: get-build-cluster-credentials
	kubectl cluster-info

get-arm-cluster-credentials: PROJECT=$(PROJECT_ARM)
get-arm-cluster-credentials: CLUSTER=$(CLUSTER_ARM)
get-arm-cluster-credentials: ZONE=$(ZONE_ARM)

deploy-arm: get-arm-cluster-credentials
	kubectl apply -f ./cluster/gke/arm

query-arm: get-arm-cluster-credentials
	kubectl cluster-info

get-private-cluster-credentials: PROJECT=$(PROJECT_PRIVATE)
get-private-cluster-credentials: CLUSTER=$(CLUSTER_PRIVATE)
get-private-cluster-credentials: ZONE=$(ZONE_PRIVATE)

deploy-private: get-private-cluster-credentials
	kubectl apply -f ./cluster/gke/private/

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

# ---------------------------------------------------------------------------
# AWS / EKS variants. Manifests live under ./cluster/eks/ and authentication
# goes through `aws eks update-kubeconfig` (see Makefile.aws.mk).
# ---------------------------------------------------------------------------

update-config-aws-dry-run: get-cluster-credentials-aws
	./recreate_prow_configmaps.py \
		--job-config-dir=cluster/eks/jobs \
		--prow-config-path=config-aws.yaml \
		--plugins-config-path=plugins.yaml

update-config-aws: get-cluster-credentials-aws
	./recreate_prow_configmaps.py \
		--job-config-dir=cluster/eks/jobs \
		--prow-config-path=config-aws.yaml \
		--plugins-config-path=plugins.yaml \
		--wet

deploy-aws: get-cluster-credentials-aws
# Apply the ProwJob CRD with --server-side=true due to its size.
	kubectl apply --server-side=true -f ./cluster/eks/prowjob-crd/
	kubectl apply -f ./cluster/eks/

deploy-gcsweb-aws: get-cluster-credentials-aws
	kubectl apply -f ./cluster/eks/gcsweb/ -n gcs

# TODO(clarketm): enable pruning
deploy-monitoring-aws: get-cluster-credentials-aws
	kubectl apply -f ./cluster/eks/monitoring/prometheus-operator-crds/
	$(MAKE) -C cluster/eks/monitoring/mixins clean apply
	kubectl apply -f ./cluster/eks/monitoring/

get-build-cluster-credentials-aws: CLUSTER=$(CLUSTER_BUILD_AWS)

deploy-build-aws: get-build-cluster-credentials-aws
	kubectl apply -f ./cluster/eks/build/

query-build-aws: get-build-cluster-credentials-aws
	kubectl cluster-info

get-private-cluster-credentials-aws: CLUSTER=$(CLUSTER_PRIVATE_AWS)

deploy-private-aws: get-private-cluster-credentials-aws
	kubectl apply -f ./cluster/eks/private/

query-private-aws: get-private-cluster-credentials-aws
	kubectl cluster-info

.PHONY: gen-private-jobs deploy deploy-gcsweb deploy-build deploy-private update-config update-config-dry-run create-istio-deps-configmap create-release-deps-configmap create-deps-configmaps
.PHONY: deploy-aws deploy-gcsweb-aws deploy-monitoring-aws deploy-build-aws deploy-private-aws query-build-aws query-private-aws update-config-aws update-config-aws-dry-run

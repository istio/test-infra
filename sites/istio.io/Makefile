# GKE variables.
PROJECT ?= istio-testing
ZONE    ?= us-west1-a
CLUSTER ?= prow

# ensures that kubectl has access to the prow cluster's config context
get-cluster-credentials:
	gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT)" --zone="$(ZONE)"

# TODO add deployment step for multi-domain certificate.
.PHONY: deploy
deploy: get-cluster-credentials
	kubectl apply -f configmap-nginx.yaml
	kubectl apply -f service.yaml
	kubectl apply -f deployment.yaml

PROJECT ?= istio-io
ZONE ?= us-west1-b
CLUSTER ?= istio-io-cluster

deploy: get-cluster-credentials
	kubectl apply -f deployment.yaml
	kubectl apply -f service.yaml

get-cluster-credentials:
	gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT)" --zone="$(ZONE)"

.PHONY: deploy get-cluster-credentials

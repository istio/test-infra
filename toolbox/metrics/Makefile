PROJECT ?= istio-testing
ZONE ?= us-west1-a
CLUSTER ?= prow
NAMESPACE ?= metrics
TAG ?= $(shell date +v%Y%m%d)-$(shell git describe --tags --always --dirty)

image:
	bazel run //toolbox/metrics/cmd:metrics

deploy:
	kubectl apply -f cmd/deployment.yaml
	kubectl apply -f cmd/service.yaml
config:

ifndef SERVICE-ACCOUNT-JSON
	echo SERVICE-ACCCOUNT-JSON must be set
	exit 1
endif
	kubectl create secret generic metrics-service-account \
		--from-file=service-account.json="$(SERVICE-ACCOUNT-JSON)" \
		-n $(NAMESPACE) --dry-run -o yaml \
		| kubectl apply -f -

get-cluster-credentials:
	gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT)" --zone="$(ZONE)"

namespace:
	kubectl create namespace $(NAMESPACE)

init: get-cluster-credentials namespace

.PHONY: image deploy config namespace get-cluster-credentials init

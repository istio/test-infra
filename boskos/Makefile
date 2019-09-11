PROJECT ?= istio-prow-build
ZONE ?= us-west1-a
CLUSTER ?= prow
NAMESPACE ?= boskos
HUB ?= gcr.io/istio-testing
TAG ?= $(shell date +v%Y%m%d)-$(shell git describe --tags --always --dirty)

activate-serviceaccount:
ifdef GOOGLE_APPLICATION_CREDENTIALS
	gcloud auth activate-service-account --key-file="$(GOOGLE_APPLICATION_CREDENTIALS)"
endif

mason-image:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cmd/mason/mason istio.io/test-infra/boskos/cmd/mason/
	docker build --no-cache -t "$(HUB)/mason:$(TAG)" cmd/mason
	docker push "$(HUB)/mason:$(TAG)"
	rm cmd/mason/mason

mason-client:
	go build -o cmd/mason_client/mason_client istio.io/test-infra/boskos/cmd/mason_client

deploy: get-cluster-credentials
	kubectl apply -Rf cluster

boskos-config: get-cluster-credentials
	kubectl create configmap boskos-config\
		--from-file=resources.yaml\
		-n $(NAMESPACE) --dry-run -o yaml\
		| kubectl apply -f -

create-serviceaccount: get-cluster-credentials
ifndef SERVICE-ACCOUNT-JSON
	echo SERVICE-ACCCOUNT-JSON must be set
	exit 1
endif
	kubectl create secret generic boskos-service-account \
		--from-file=service-account.json="$(SERVICE-ACCOUNT-JSON)" \
		-n $(NAMESPACE) --dry-run -o yaml \
		| kubectl apply -f -

get-cluster-credentials: activate-serviceaccount
	gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT)" --zone="$(ZONE)"

namespace: get-cluster-credentials
	kubectl create namespace $(NAMESPACE)

init: namespace create-serviceaccount

.PHONY: mason-image mason-client deploy boskos-config activate-serviceaccount create-serviceaccount get-cluster-credentials namespace init

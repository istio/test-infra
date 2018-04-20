PROJECT ?= istio-testing
ZONE ?= us-west1-a
CLUSTER ?= prow
NAMESPACE ?= boskos
HUB ?= gcr.io/istio-testing
TAG ?= $(shell date +v%Y%m%d)-$(shell git describe --tags --always --dirty)

mason-image:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cmd/mason/mason istio.io/test-infra/boskos/cmd/mason/
	docker build --no-cache -t "$(HUB)/mason:$(TAG)" cmd/mason
	docker push "$(HUB)/mason:$(TAG)"
	rm cmd/mason/mason

mason-client:
	go build -o cmd/mason_client/mason_client istio.io/test-infra/boskos/cmd/mason_client

deploy: boskos-deployment reaper-deployment janitor-deployment mason-deployment boskos-service

boskos-deployment:
	kubectl apply -f boskos-deployment.yaml

reaper-deployment:
	kubectl apply -f reaper-deployment.yaml

janitor-deployment:
	kubectl apply -f janitor-deployment.yaml

mason-deployment:
	kubectl apply -f mason-deployment.yaml

boskos-service:
	kubectl apply -f boskos-service.yaml

config: serviceaccount boskos-config mason-config

boskos-config:
	kubectl create configmap boskos-config\
		--from-file=resources.yaml\
		-n $(NAMESPACE) --dry-run -o yaml\
		| kubectl apply -f -

mason-config:
	kubectl create configmap mason-config\
		--from-file=configs.yaml\
		-n $(NAMESPACE) --dry-run -o yaml\
		| kubectl apply -f -

serviceaccount:

ifndef SERVICE-ACCOUNT-JSON
	echo SERVICE-ACCCOUNT-JSON must be set
	exit 1
endif
	kubectl create secret generic boskos-service-account \
		--from-file=service-account.json="$(SERVICE-ACCOUNT-JSON)" \
		-n $(NAMESPACE) --dry-run -o yaml \
		| kubectl apply -f -

get-cluster-credentials:
	gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT)" --zone="$(ZONE)"

namespace:
	kubectl create namespace $(NAMESPACE)

init: get-cluster-credentials namespace

.PHONY: mason-image boskos-deployment reaper-deployment janitor-deployment mason-deployment boskos-service mason-config boskos-config get-cluster-credentials namespace serviceaccount deploy init config

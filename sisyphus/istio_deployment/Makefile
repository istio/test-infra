PROJECT = istio-testing
VERSION = 0.1.0
ZONE    ?= us-west1-a
CLUSTER ?= istio-sisyphus
MODIFIED_YAML := "/tmp/local.sisyphus-deployment.yaml"

default: image push

image:
	docker build -t "gcr.io/$(PROJECT)/sisyphus:$(VERSION)" -f Dockerfile .

push:
	gcloud docker -- push "gcr.io/$(PROJECT)/sisyphus:$(VERSION)"

get-cluster-credentials:
	gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT)" --zone="$(ZONE)"

deploy: get-cluster-credentials
	sed 's/image: gcr.io\/istio-testing\/sisyphus:.*/image: gcr.io\/${PROJECT}\/sisyphus:'"${VERSION}"'/g' \
	sisyphus-deployment.yaml > ${MODIFIED_YAML}
	cat ${MODIFIED_YAML}
	kubectl apply -f ${MODIFIED_YAML}

stop: get-cluster-credentials
	kubectl delete deployment sisyphus

update: get-cluster-credentials stop deploy

run:
	docker run gcr.io/$(PROJECT)/sisyphus:$(VERSION)

# Open shell to running prowbazel image for experimentation.
run.bash:
	docker run -it --entrypoint bash --privileged gcr.io/$(PROJECT)/sisyphus:$(VERSION)

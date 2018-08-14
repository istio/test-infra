PROJECT = istio-testing
VERSION = 0.5.7

# Note: The build directory is the root of the istio/test-infra repository, not ./
image:
	docker build -t istio-builder -f istio_builder/Dockerfile ../../
	docker tag istio-builder gcr.io/$(PROJECT)/istio-builder:$(VERSION)
	docker tag istio-builder gcr.io/$(PROJECT)/istio-builder:latest
	docker build -t gcr.io/$(PROJECT)/prow-builder:$(VERSION) -f prow_builder/Dockerfile ../..


push:
	docker push "gcr.io/$(PROJECT)/istio-builder:latest"
	docker push "gcr.io/$(PROJECT)/istio-builder:$(VERSION)"
	docker push "gcr.io/$(PROJECT)/prow-builder:$(VERSION)"


.PHONY: image push

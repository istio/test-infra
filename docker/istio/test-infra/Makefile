PROJECT ?= istio-testing
VERSION ?= 0.6
IMAGE ?= test-infra-builder

# Note: The build directory is the root of the istio/test-infra repository, not ./
image:
	docker build --cache-from gcr.io/$(PROJECT)/$(IMAGE):latest \
		-t gcr.io/$(PROJECT)/$(IMAGE):$(VERSION) \
		-t gcr.io/$(PROJECT)/$(IMAGE):latest .


push:
	docker push gcr.io/$(PROJECT)/$(IMAGE):latest
	docker push gcr.io/$(PROJECT)/$(IMAGE):$(VERSION)


.PHONY: image push

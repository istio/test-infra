PROJECT = istio-io
VERSION = 0.3.0

# Note: The build directory is the root of the istio/test-infra repository, not ./
image:
	docker build -t "gcr.io/$(PROJECT)/istio-builder:$(VERSION)" -f Dockerfile ../../

push:
	gcloud docker -- push "gcr.io/$(PROJECT)/istio-builder:$(VERSION)"

# Open shell to running istio-builder image for experimentation.
run:
	docker run -it --entrypoint bash --privileged gcr.io/$(PROJECT)/istio-builder:$(VERSION)

.PHONY: image push

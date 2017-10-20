PROJECT = istio-testing
VERSION = 0.1.0

# Note: The build directory is the root of the istio/test-infra repository, not ./
image:
	docker build -t "gcr.io/$(PROJECT)/prowbuilder:$(VERSION)" -f Dockerfile ../../

push:
	gcloud docker -- push "gcr.io/$(PROJECT)/prowbuilder:$(VERSION)"

# Open shell to running prowbuilder image for experimentation.
run:
	docker run -it --entrypoint bash --privileged gcr.io/$(PROJECT)/prowbuilder:$(VERSION)

.PHONY: image push

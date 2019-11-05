PROJECT = istio-testing
VERSION ?= 0.5.16

# Note: The build directory is the root of the istio/test-infra repository, not ./
image:
	docker build -t "gcr.io/$(PROJECT)/prowbazel:$(VERSION)" -f Dockerfile ../../

# Push image only if it does not exist.
push-safe:
	DOCKER_CLI_EXPERIMENTAL=enabled docker manifest inspect "gcr.io/$(PROJECT)/prowbazel:$(VERSION)" &>/dev/null || $(MAKE) push

push:
	docker push "gcr.io/$(PROJECT)/prowbazel:$(VERSION)"

# Open shell to running prowbazel image for experimentation.
run:
	docker run -it --entrypoint bash --privileged gcr.io/$(PROJECT)/prowbazel:$(VERSION)
# Example of running a job:
# ~/bootstrap.py --repo="github.com/<org>/<repo>" --job="<job-name>" --clean --root="${GOPATH}/src" --pull=master:<sha>,<pr-#>:<sha>

.PHONY: image push push-safe

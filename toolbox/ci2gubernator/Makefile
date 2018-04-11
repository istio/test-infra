REPO_ROOT := $(shell git rev-parse --show-toplevel)

default: push

build:
	bazel build //toolbox/ci2gubernator

push: build
	# upload artifacts to GCS
	gsutil cp \
	$(REPO_ROOT)/bazel-bin/toolbox/ci2gubernator/linux_amd64_stripped/ci2gubernator \
	gs://istio-tools/ci2gubernator/
	# make it public
	gsutil acl ch -u AllUsers:R gs://istio-tools/ci2gubernator/ci2gubernator

.PHONY: default build push

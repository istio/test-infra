default: push

build:
	bazel build //toolbox/ci_to_gubernator

push: build
	# upload artifacts to GCS
	gsutil cp \
	../../bazel-bin/toolbox/ci_to_gubernator/linux_amd64_stripped/ci_to_gubernator \
	gs://istio-tools/ci_to_gubernator
	# make it public
	gsutil acl ch -u AllUsers:R gs://istio-tools/ci_to_gubernator/ci_to_gubernator

.PHONY: default build push

workspace(name = "com_github_istio_test_infra")

git_repository(
    name = "io_bazel_rules_go",
    commit = "9cf23e2aab101f86e4f51d8c5e0f14c012c2161c",  # Oct 12, 2017
    remote = "https://github.com/bazelbuild/rules_go.git",
)

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_rules_dependencies",
    "go_repository",
    "go_register_toolchains",
)

go_rules_dependencies()

go_register_toolchains(go_version = "1.8.3")

load("@io_bazel_rules_go//proto:def.bzl", "proto_register_toolchains")

proto_register_toolchains()

##
## docker
##

git_repository(
    name = "io_bazel_rules_docker",
    commit = "9dd92c73e7c8cf07ad5e0dca89a3c3c422a3ab7d",  # Sep 27, 2017
    remote = "https://github.com/bazelbuild/rules_docker.git",
)

go_repository(
    name = "com_github_docker_distribution",
    commit = "a25b9ef0c9fe242ac04bb20d3a028442b7d266b6",  # Apr 5, 2017 (v2.6.1)
    importpath = "github.com/docker/distribution",
)

load("@io_bazel_rules_docker//docker:docker.bzl", "docker_repositories", "docker_pull")

docker_repositories()

docker_pull(
    name = "distroless",
    registry = "gcr.io",
    repository = "distroless/base",
    tag = "latest",
)

# Vendors
#

load("//:go_vendor_repositories.bzl", "go_vendor_repositories")

go_vendor_repositories()

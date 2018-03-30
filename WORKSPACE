workspace(name = "com_github_istio_test_infra")

git_repository(
    name = "bazel_skylib",
    commit = "2169ae1c374aab4a09aa90e65efe1a3aad4e279b",
    remote = "https://github.com/bazelbuild/bazel-skylib.git",
)

load("@bazel_skylib//:lib.bzl", "versions")

git_repository(
    name = "io_bazel_rules_go",
    commit = "bdf2df58c0d352ffa262ae4b36c7a1a2d6e3f0c9",
    remote = "https://github.com/bazelbuild/rules_go.git",
)

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_rules_dependencies",
    "go_repository",
    "go_register_toolchains",
)

go_rules_dependencies()

go_register_toolchains(go_version = "1.9.3")

load("@io_bazel_rules_go//proto:def.bzl", "proto_register_toolchains")

##
## docker
##

# You *must* import the Go rules before setting up the go_image rules.
git_repository(
    name = "io_bazel_rules_docker",
    remote = "https://github.com/bazelbuild/rules_docker.git",
    tag = "v0.4.0",
)

load(
    "@io_bazel_rules_docker//container:container.bzl",
    "container_push",
    container_repositories = "repositories",
)
load(
    "@io_bazel_rules_docker//go:image.bzl",
    _go_image_repos = "repositories",
)

container_repositories()

_go_image_repos()

# Vendors
#

load("//:go_vendor_repositories.bzl", "go_vendor_repositories")

go_vendor_repositories()

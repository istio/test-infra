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

go_repository(
    name = "com_google_cloud_go",
    commit = "2d3a6656c17a60b0815b7e06ab0be04eacb6e613",
    importpath = "cloud.google.com/go",
)

go_repository(
    name = "com_github_beorn7_perks",
    commit = "4c0e84591b9aa9e6dcfdf3e020114cd81f89d5f9",
    importpath = "github.com/beorn7/perks",
)

go_repository(
    name = "com_github_golang_protobuf",
    commit = "1643683e1b54a9e88ad26d98f81400c8c9d9f4f9",
    importpath = "github.com/golang/protobuf",
)

go_repository(
    name = "com_github_google_go_github",
    commit = "8c08f4fba5e05e0fd2821a5f80cf0cf643bd5314",
    importpath = "github.com/google/go-github",
)

go_repository(
    name = "com_github_google_go_querystring",
    commit = "53e6ce116135b80d037921a7fdd5138cf32d7a8a",
    importpath = "github.com/google/go-querystring",
)

go_repository(
    name = "com_github_googleapis_gax_go",
    commit = "317e0006254c44a0ac427cc52a0e083ff0b9622f",
    importpath = "github.com/googleapis/gax-go",
)

go_repository(
    name = "com_github_hashicorp_errwrap",
    commit = "7554cd9344cec97297fa6649b055a8c98c2a1e55",
    importpath = "github.com/hashicorp/errwrap",
)

go_repository(
    name = "com_github_hashicorp_go_multierror",
    commit = "83588e72410abfbe4df460eeb6f30841ae47d4c4",
    importpath = "github.com/hashicorp/go-multierror",
)

go_repository(
    name = "com_github_matttproud_golang_protobuf_extensions",
    commit = "3247c84500bff8d9fb6d579d800f20b3e091582c",
    importpath = "github.com/matttproud/golang_protobuf_extensions",
)

go_repository(
    name = "com_github_prometheus_client_golang",
    commit = "c5b7fccd204277076155f10851dad72b76a49317",
    importpath = "github.com/prometheus/client_golang",
)

go_repository(
    name = "com_github_prometheus_client_model",
    commit = "6f3806018612930941127f2a7c6c453ba2c527d2",
    importpath = "github.com/prometheus/client_model",
)

go_repository(
    name = "com_github_prometheus_common",
    commit = "e3fb1a1acd7605367a2b378bc2e2f893c05174b7",
    importpath = "github.com/prometheus/common",
)

go_repository(
    name = "com_github_prometheus_procfs",
    commit = "a6e9df898b1336106c743392c48ee0b71f5c4efa",
    importpath = "github.com/prometheus/procfs",
)

go_repository(
    name = "org_golang_x_net",
    commit = "a337091b0525af65de94df2eb7e98bd9962dcbe2",
    importpath = "golang.org/x/net",
)

go_repository(
    name = "org_golang_x_oauth2",
    commit = "9ff8ebcc8e241d46f52ecc5bff0e5a2f2dbef402",
    importpath = "golang.org/x/oauth2",
)

go_repository(
    name = "org_golang_x_text",
    commit = "88f656faf3f37f690df1a32515b479415e1a6769",
    importpath = "golang.org/x/text",
)

go_repository(
    name = "org_golang_google_api",
    commit = "34bc9e3ce3692fb01d22f685df28518773cbdae9",
    importpath = "google.golang.org/api",
)

go_repository(
    name = "org_golang_google_appengine",
    commit = "150dc57a1b433e64154302bdc40b6bb8aefa313a",
    importpath = "google.golang.org/appengine",
)

go_repository(
    name = "org_golang_google_genproto",
    commit = "11c7f9e547da6db876260ce49ea7536985904c9b",
    importpath = "google.golang.org/genproto",
)

go_repository(
    name = "org_golang_google_grpc",
    commit = "5ffe3083946d5603a0578721101dc8165b1d5b5f",
    importpath = "google.golang.org/grpc",
)

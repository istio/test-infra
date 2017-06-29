load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
load("@io_bazel_rules_docker//docker:docker.bzl", "docker_build")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    visibility = ["//visibility:private"],
    deps = [
        "@com_github_prometheus_client_golang//prometheus:go_default_library",
        "@com_github_prometheus_client_golang//prometheus/promhttp:go_default_library",
        "@com_google_cloud_go//storage:go_default_library",
        "@org_golang_x_net//context:go_default_library",
    ],
)

go_binary(
    name = "fetcher",
    gc_linkopts = [
        "-linkmode",
        "external",
        "-extldflags",
        "-static",
    ],
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)

docker_build(
    name = "metrics_fetcher",
    base = "@distroless//image:image.tar",
    cmd = ["./fetcher"],
    files = [":fetcher"],
)

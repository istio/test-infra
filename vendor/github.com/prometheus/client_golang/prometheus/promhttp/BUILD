load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["http.go"],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_prometheus_client_golang//prometheus:go_default_library",
        "@com_github_prometheus_common//expfmt:go_default_library",
    ],
)

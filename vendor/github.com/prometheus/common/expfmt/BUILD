load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "decode.go",
        "encode.go",
        "expfmt.go",
        "text_create.go",
        "text_parse.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_golang_protobuf//proto:go_default_library",
        "@com_github_matttproud_golang_protobuf_extensions//pbutil:go_default_library",
        "@com_github_prometheus_client_model//go:go_default_library",
        "@com_github_prometheus_common//internal/bitbucket.org/ww/goautoneg:go_default_library",
        "@com_github_prometheus_common//model:go_default_library",
    ],
)

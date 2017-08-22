load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["dial.go"],
    visibility = ["//visibility:public"],
    deps = [
        "@org_golang_google_api//internal:go_default_library",
        "@org_golang_google_api//option:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_google_grpc//credentials:go_default_library",
        "@org_golang_google_grpc//credentials/oauth:go_default_library",
        "//vendor/golang.org.x/net/context:go_default_library",
    ],
)

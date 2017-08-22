load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["option.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/golang.org/x/oauth2:go_default_library",
        "//vendor/google.golang.org/api/internal:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)

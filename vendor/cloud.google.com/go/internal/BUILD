load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["retry.go"],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_googleapis_gax_go//:go_default_library",
        "//vendor/golang.org.x/net/context:go_default_library",
    ],
)

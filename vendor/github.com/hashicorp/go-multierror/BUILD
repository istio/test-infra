load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "append.go",
        "flatten.go",
        "format.go",
        "multierror.go",
        "prefix.go",
    ],
    visibility = ["//visibility:public"],
    deps = ["@com_github_hashicorp_errwrap//:go_default_library"],
)

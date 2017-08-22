load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "googleapi.go",
        "types.go",
    ],
    visibility = ["//visibility:public"],
    deps = ["//vendor/google.golang.org/api/googleapi/internal/uritemplates:go_default_library"],
)

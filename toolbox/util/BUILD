load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "commonUtils.go",
        "githubUtils.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_google_go_github//github:go_default_library",
    ],
)

load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "goimports.go",
        "goimports_gc.go",
    ],
    visibility = ["//visibility:private"],
    deps = ["//vendor/golang.org/x/tools/imports:go_default_library"],
)

go_binary(
    name = "goimports",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)

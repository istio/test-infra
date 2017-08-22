load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "call_option.go",
        "gax.go",
        "header.go",
        "invoke.go",
        "path_template.go",
        "path_template_parser.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_google_grpc//codes:go_default_library",
        "//vendor/golang.org.x/net/context:go_default_library",
    ],
)

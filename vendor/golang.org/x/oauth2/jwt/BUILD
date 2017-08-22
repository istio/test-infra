load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["jwt.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/golang.org.x/net/context:go_default_library",
        "@org_golang_x_oauth2//:go_default_library",
        "@org_golang_x_oauth2//internal:go_default_library",
        "@org_golang_x_oauth2//jws:go_default_library",
    ],
)

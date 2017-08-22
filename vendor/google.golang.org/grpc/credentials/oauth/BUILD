load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["oauth.go"],
    visibility = ["//visibility:public"],
    deps = [
        "@org_golang_google_grpc//credentials:go_default_library",
        "//vendor/golang.org.x/net/context:go_default_library",
        "@org_golang_x_oauth2//:go_default_library",
        "@org_golang_x_oauth2//google:go_default_library",
        "@org_golang_x_oauth2//jwt:go_default_library",
    ],
)

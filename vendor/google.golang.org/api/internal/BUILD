load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "creds.go",
        "pool.go",
        "settings.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_google_grpc//naming:go_default_library",
        "//vendor/golang.org.x/net/context:go_default_library",
        "@org_golang_x_oauth2//:go_default_library",
        "@org_golang_x_oauth2//google:go_default_library",
    ],
)

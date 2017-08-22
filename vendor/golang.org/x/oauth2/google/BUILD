load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "appengine.go",
        "default.go",
        "google.go",
        "jwt.go",
        "sdk.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@com_google_cloud_go//compute/metadata:go_default_library",
        "//vendor/golang.org.x/net/context:go_default_library",
        "@org_golang_x_oauth2//:go_default_library",
        "@org_golang_x_oauth2//internal:go_default_library",
        "@org_golang_x_oauth2//jws:go_default_library",
        "@org_golang_x_oauth2//jwt:go_default_library",
    ],
)

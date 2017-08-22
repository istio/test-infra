load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "appengine.go",
        "appengine_vm.go",
        "errors.go",
        "identity.go",
        "namespace.go",
        "timeout.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_golang_protobuf//proto:go_default_library",
        "@org_golang_google_appengine//internal:go_default_library",
        "@org_golang_google_appengine//internal/app_identity:go_default_library",
        "@org_golang_google_appengine//internal/modules:go_default_library",
        "//vendor/golang.org.x/net/context:go_default_library",
    ],
)

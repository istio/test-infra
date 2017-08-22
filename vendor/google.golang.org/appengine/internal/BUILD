load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "api.go",
        "api_common.go",
        "app_id.go",
        "identity.go",
        "identity_vm.go",
        "internal.go",
        "main_vm.go",
        "metadata.go",
        "net.go",
        "transaction.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_golang_protobuf//proto:go_default_library",
        "@org_golang_google_appengine//internal/base:go_default_library",
        "@org_golang_google_appengine//internal/datastore:go_default_library",
        "@org_golang_google_appengine//internal/log:go_default_library",
        "@org_golang_google_appengine//internal/remote_api:go_default_library",
        "//vendor/golang.org.x/net/context:go_default_library",
    ],
)

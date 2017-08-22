load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["storage-gen.go"],
    visibility = ["//visibility:public"],
    deps = [
        "@org_golang_google_api//gensupport:go_default_library",
        "@org_golang_google_api//googleapi:go_default_library",
        "//vendor/golang.org.x/net/context:go_default_library",
        "//vendor/golang.org.x/net/context/ctxhttp:go_default_library",
    ],
)

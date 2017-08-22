load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "backoff.go",
        "buffer.go",
        "doc.go",
        "header.go",
        "json.go",
        "jsonfloat.go",
        "media.go",
        "params.go",
        "resumable.go",
        "retry.go",
        "send.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@org_golang_google_api//googleapi:go_default_library",
        "//vendor/golang.org.x/net/context:go_default_library",
        "//vendor/golang.org.x/net/context/ctxhttp:go_default_library",
    ],
)

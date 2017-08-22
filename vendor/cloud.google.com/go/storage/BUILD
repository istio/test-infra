load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "acl.go",
        "bucket.go",
        "copy.go",
        "doc.go",
        "iam.go",
        "invoke.go",
        "reader.go",
        "storage.go",
        "writer.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_googleapis_gax_go//:go_default_library",
        "@com_google_cloud_go//iam:go_default_library",
        "@com_google_cloud_go//internal:go_default_library",
        "@com_google_cloud_go//internal/optional:go_default_library",
        "@com_google_cloud_go//internal/version:go_default_library",
        "@org_golang_google_api//googleapi:go_default_library",
        "@org_golang_google_api//iterator:go_default_library",
        "@org_golang_google_api//option:go_default_library",
        "@org_golang_google_api//storage/v1:go_default_library",
        "@org_golang_google_api//transport:go_default_library",
        "@org_golang_google_genproto//googleapis/iam/v1:go_default_library",
        "//vendor/golang.org.x/net/context:go_default_library",
    ],
)

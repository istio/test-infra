load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "control.go",
        "go17.go",
        "handler_server.go",
        "http2_client.go",
        "http2_server.go",
        "http_util.go",
        "transport.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@org_golang_google_grpc//codes:go_default_library",
        "@org_golang_google_grpc//credentials:go_default_library",
        "@org_golang_google_grpc//grpclog:go_default_library",
        "@org_golang_google_grpc//keepalive:go_default_library",
        "@org_golang_google_grpc//metadata:go_default_library",
        "@org_golang_google_grpc//peer:go_default_library",
        "@org_golang_google_grpc//stats:go_default_library",
        "@org_golang_google_grpc//tap:go_default_library",
        "//vendor/golang.org.x/net/context:go_default_library",
        "//vendor/golang.org.x/net/http2:go_default_library",
        "//vendor/golang.org.x/net/http2/hpack:go_default_library",
    ],
)

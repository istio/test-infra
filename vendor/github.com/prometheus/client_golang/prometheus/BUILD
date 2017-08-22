load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "collector.go",
        "counter.go",
        "desc.go",
        "doc.go",
        "expvar_collector.go",
        "fnv.go",
        "gauge.go",
        "go_collector.go",
        "histogram.go",
        "http.go",
        "metric.go",
        "process_collector.go",
        "registry.go",
        "summary.go",
        "untyped.go",
        "value.go",
        "vec.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_beorn7_perks//quantile:go_default_library",
        "@com_github_golang_protobuf//proto:go_default_library",
        "@com_github_prometheus_client_model//go:go_default_library",
        "@com_github_prometheus_common//expfmt:go_default_library",
        "@com_github_prometheus_common//model:go_default_library",
        "@com_github_prometheus_procfs//:go_default_library",
    ],
)

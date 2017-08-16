load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
    	"main.go",
    	"dependency.go",
    ],
    deps = [
        "//toolbox/util:go_default_library",
        "@com_github_google_go_github//github:go_default_library",
    ],
)

go_binary(
    name = "update_deps",
    library = ":go_default_library",
)

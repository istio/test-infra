package(default_visibility = ["//visibility:public"])

sh_binary(
    name = "format",
    srcs = ["format.sh"],
    data = [
        "@com_github_bazelbuild_buildtools//buildifier",
        "@io_bazel_rules_go_toolchain//:toolchain",
        "@org_golang_x_tools//cmd/goimports",
    ],
)

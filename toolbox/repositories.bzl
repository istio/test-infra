load("@io_bazel_rules_go//go:def.bzl", "new_go_repository")

def toolbox_repositories():

  # Format tool
  native.git_repository(
      name = "com_github_bazelbuild_buildtools",
      remote = "https://github.com/bazelbuild/buildtools.git",
      tag = "0.4.5",
  )

  new_go_repository(
      name = "org_golang_x_tools",
      commit = "3d92dd60033c312e3ae7cac319c792271cf67e37",
      importpath = "golang.org/x/tools",
  )

  # Pkg Check
  new_go_repository(
      name = "org_golang_google_api",
      commit = "48e49d1645e228d1c50c3d54fb476b2224477303",
      importpath = "google.golang.org/api",
  )

  new_go_repository(
      name = "com_github_googleapis_gax_go",
      commit = "9af46dd5a1713e8b5cd71106287eba3cefdde50b",
      importpath = "github.com/googleapis/gax-go",
  )

  new_go_repository(
      name = "org_golang_google_genproto",
      commit = "411e09b969b1170a9f0c467558eb4c4c110d9c77",
      importpath = "google.golang.org/genproto",
  )

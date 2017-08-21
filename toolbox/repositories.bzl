load("@io_bazel_rules_go//go:def.bzl", "go_repository")

def toolbox_repositories():

  # Format tool
  go_repository(
      name = "org_golang_x_tools",
      commit = "3d92dd60033c312e3ae7cac319c792271cf67e37",
      importpath = "golang.org/x/tools",
  )

  # Pkg Check
  go_repository(
      name = "org_golang_google_api",
      commit = "48e49d1645e228d1c50c3d54fb476b2224477303",
      importpath = "google.golang.org/api",
  )

  go_repository(
      name = "org_golang_google_genproto",
      commit = "411e09b969b1170a9f0c467558eb4c4c110d9c77",
      importpath = "google.golang.org/genproto",
  )

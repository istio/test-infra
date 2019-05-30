workspace(name = "io_istio_test_infra")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")

# buildifier is written in Go and hence needs rules_go to be built.
# See https://github.com/bazelbuild/rules_go for the up to date setup instructions.
http_archive(
    name = "io_bazel_rules_go",
    sha256 = "a82a352bffae6bee4e95f68a8d80a70e87f42c4741e6a448bec11998fcc82329",
    urls = ["https://github.com/bazelbuild/rules_go/releases/download/0.18.5/rules_go-0.18.5.tar.gz"],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "3c681998538231a2d24d0c07ed5a7658cb72bfb5fd4bf9911157c0e9ac6a2687",
    urls = ["https://github.com/bazelbuild/bazel-gazelle/releases/download/0.17.0/bazel-gazelle-0.17.0.tar.gz"],
)

http_archive(
    name = "com_github_bazelbuild_buildtools",
    strip_prefix = "buildtools-0.20.0",
    url = "https://github.com/bazelbuild/buildtools/archive/0.20.0.zip",
)

http_archive(
    name = "com_github_atlassian_bazel_tools",
    strip_prefix = "bazel-tools-864fde1c98ab943cf8bc61768bff5473d1277068",
    urls = ["https://github.com/atlassian/bazel-tools/archive/864fde1c98ab943cf8bc61768bff5473d1277068.zip"],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(go_version = "1.12.5")

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

load("@com_github_bazelbuild_buildtools//buildifier:deps.bzl", "buildifier_dependencies")

buildifier_dependencies()

load("@com_github_atlassian_bazel_tools//golangcilint:deps.bzl", "golangcilint_dependencies")

golangcilint_dependencies()
##
## docker
##

# You *must* import the Go rules before setting up the go_image rules.
http_archive(
    name = "io_bazel_rules_docker",
    sha256 = "aed1c249d4ec8f703edddf35cbe9dfaca0b5f5ea6e4cd9e83e99f3b0d1136c3d",
    strip_prefix = "rules_docker-0.7.0",
    urls = ["https://github.com/bazelbuild/rules_docker/archive/v0.7.0.tar.gz"],
)

load(
    "@io_bazel_rules_docker//go:image.bzl",
    _go_image_repos = "repositories",
)

_go_image_repos()

go_repository(
    name = "com_github_beorn7_perks",
    commit = "4c0e84591b9a",
    importpath = "github.com/beorn7/perks",
)

go_repository(
    name = "com_github_bwmarrin_snowflake",
    commit = "68117e6bbede",
    importpath = "github.com/bwmarrin/snowflake",
)

go_repository(
    name = "com_github_davecgh_go_spew",
    importpath = "github.com/davecgh/go-spew",
    tag = "v1.1.1",
)

go_repository(
    name = "com_github_deckarep_golang_set",
    commit = "1d4478f51bed",
    importpath = "github.com/deckarep/golang-set",
)

go_repository(
    name = "com_github_fsnotify_fsnotify",
    importpath = "github.com/fsnotify/fsnotify",
    tag = "v1.4.7",
)

go_repository(
    name = "com_github_ghodss_yaml",
    importpath = "github.com/ghodss/yaml",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_go_yaml_yaml",
    importpath = "github.com/go-yaml/yaml",
    tag = "v2.1.0",
)

go_repository(
    name = "com_github_gogo_protobuf",
    build_file_proto_mode = "disable_global",
    importpath = "github.com/gogo/protobuf",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_golang_glog",
    commit = "23def4e6c14b",
    importpath = "github.com/golang/glog",
)

go_repository(
    name = "com_github_golang_lint",
    commit = "85993ffd0a6c",
    importpath = "github.com/golang/lint",
)

go_repository(
    name = "com_github_golang_protobuf",
    importpath = "github.com/golang/protobuf",
    tag = "v1.3.1",
)

go_repository(
    name = "com_github_google_btree",
    commit = "4030bb1f1f0c",
    importpath = "github.com/google/btree",
)

go_repository(
    name = "com_github_google_go_cmp",
    importpath = "github.com/google/go-cmp",
    tag = "v0.2.0",
)

go_repository(
    name = "com_github_google_go_github",
    commit = "8c08f4fba5e0",
    importpath = "github.com/google/go-github",
)

go_repository(
    name = "com_github_google_go_querystring",
    commit = "53e6ce116135",
    importpath = "github.com/google/go-querystring",
)

go_repository(
    name = "com_github_google_gofuzz",
    commit = "24818f796faf",
    importpath = "github.com/google/gofuzz",
)

go_repository(
    name = "com_github_googleapis_gax_go",
    importpath = "github.com/googleapis/gax-go",
    tag = "v2.0.0",
)

go_repository(
    name = "com_github_googleapis_gnostic",
    importpath = "github.com/googleapis/gnostic",
    tag = "v0.2.0",
)

go_repository(
    name = "com_github_gorilla_context",
    commit = "1ea25387ff6f",
    importpath = "github.com/gorilla/context",
)

go_repository(
    name = "com_github_gorilla_securecookie",
    importpath = "github.com/gorilla/securecookie",
    tag = "v1.1.1",
)

go_repository(
    name = "com_github_gorilla_sessions",
    commit = "ca9ada445741",
    importpath = "github.com/gorilla/sessions",
)

go_repository(
    name = "com_github_gregjones_httpcache",
    commit = "c63ab54fda8f",
    importpath = "github.com/gregjones/httpcache",
)

go_repository(
    name = "com_github_hashicorp_errwrap",
    commit = "7554cd9344ce",
    importpath = "github.com/hashicorp/errwrap",
)

go_repository(
    name = "com_github_hashicorp_go_multierror",
    commit = "83588e72410a",
    importpath = "github.com/hashicorp/go-multierror",
)

go_repository(
    name = "com_github_hpcloud_tail",
    importpath = "github.com/hpcloud/tail",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_imdario_mergo",
    importpath = "github.com/imdario/mergo",
    tag = "v0.3.6",
)

go_repository(
    name = "com_github_json_iterator_go",
    importpath = "github.com/json-iterator/go",
    tag = "v1.1.5",
)

go_repository(
    name = "com_github_knative_build",
    importpath = "github.com/knative/build",
    tag = "v0.1.0",
)

go_repository(
    name = "com_github_kr_pretty",
    importpath = "github.com/kr/pretty",
    tag = "v0.1.0",
)

go_repository(
    name = "com_github_kr_pty",
    importpath = "github.com/kr/pty",
    tag = "v1.1.1",
)

go_repository(
    name = "com_github_kr_text",
    importpath = "github.com/kr/text",
    tag = "v0.1.0",
)

go_repository(
    name = "com_github_mattn_go_zglob",
    commit = "c436403c742d",
    importpath = "github.com/mattn/go-zglob",
)

go_repository(
    name = "com_github_matttproud_golang_protobuf_extensions",
    importpath = "github.com/matttproud/golang_protobuf_extensions",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_modern_go_concurrent",
    commit = "bacd9c7ef1dd",
    importpath = "github.com/modern-go/concurrent",
)

go_repository(
    name = "com_github_modern_go_reflect2",
    importpath = "github.com/modern-go/reflect2",
    tag = "v1.0.1",
)

go_repository(
    name = "com_github_onsi_ginkgo",
    importpath = "github.com/onsi/ginkgo",
    tag = "v1.8.0",
)

go_repository(
    name = "com_github_onsi_gomega",
    importpath = "github.com/onsi/gomega",
    tag = "v1.5.0",
)

go_repository(
    name = "com_github_peterbourgon_diskv",
    importpath = "github.com/peterbourgon/diskv",
    tag = "v2.0.1",
)

go_repository(
    name = "com_github_pmezard_go_difflib",
    importpath = "github.com/pmezard/go-difflib",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_prometheus_client_golang",
    importpath = "github.com/prometheus/client_golang",
    tag = "v0.8.0",
)

go_repository(
    name = "com_github_prometheus_client_model",
    commit = "6f3806018612",
    importpath = "github.com/prometheus/client_model",
)

go_repository(
    name = "com_github_prometheus_common",
    commit = "e3fb1a1acd76",
    importpath = "github.com/prometheus/common",
)

go_repository(
    name = "com_github_prometheus_procfs",
    commit = "a6e9df898b13",
    importpath = "github.com/prometheus/procfs",
)

go_repository(
    name = "com_github_satori_go_uuid",
    importpath = "github.com/satori/go.uuid",
    tag = "v1.2.0",
)

go_repository(
    name = "com_github_shurcool_githubv4",
    commit = "51d7b505e2e9",
    importpath = "github.com/shurcooL/githubv4",
)

go_repository(
    name = "com_github_shurcool_go",
    commit = "47fa5b7ceee6",
    importpath = "github.com/shurcooL/go",
)

go_repository(
    name = "com_github_shurcool_graphql",
    commit = "3d276b9dcc6b",
    importpath = "github.com/shurcooL/graphql",
)

go_repository(
    name = "com_github_sirupsen_logrus",
    importpath = "github.com/sirupsen/logrus",
    tag = "v1.0.4",
)

go_repository(
    name = "com_github_spf13_pflag",
    importpath = "github.com/spf13/pflag",
    tag = "v1.0.1",
)

go_repository(
    name = "com_github_stretchr_objx",
    importpath = "github.com/stretchr/objx",
    tag = "v0.1.0",
)

go_repository(
    name = "com_github_stretchr_testify",
    importpath = "github.com/stretchr/testify",
    tag = "v1.3.0",
)

go_repository(
    name = "com_google_cloud_go",
    importpath = "cloud.google.com/go",
    tag = "v0.23.0",
)

go_repository(
    name = "in_gopkg_airbrake_gobrake_v2",
    importpath = "gopkg.in/airbrake/gobrake.v2",
    tag = "v2.0.9",
)

go_repository(
    name = "in_gopkg_check_v1",
    commit = "788fd7840127",
    importpath = "gopkg.in/check.v1",
)

go_repository(
    name = "in_gopkg_fsnotify_v1",
    importpath = "gopkg.in/fsnotify.v1",
    tag = "v1.4.7",
)

go_repository(
    name = "in_gopkg_gemnasium_logrus_airbrake_hook_v2",
    importpath = "gopkg.in/gemnasium/logrus-airbrake-hook.v2",
    tag = "v2.1.2",
)

go_repository(
    name = "in_gopkg_inf_v0",
    importpath = "gopkg.in/inf.v0",
    tag = "v0.9.1",
)

go_repository(
    name = "in_gopkg_robfig_cron_v2",
    commit = "be2e0b0deed5",
    importpath = "gopkg.in/robfig/cron.v2",
)

go_repository(
    name = "in_gopkg_tomb_v1",
    commit = "dd632973f1e7",
    importpath = "gopkg.in/tomb.v1",
)

go_repository(
    name = "in_gopkg_yaml_v2",
    importpath = "gopkg.in/yaml.v2",
    tag = "v2.2.2",
)

go_repository(
    name = "io_k8s_api",
    build_file_proto_mode = "disable_global",
    commit = "173ce66c1e39",
    importpath = "k8s.io/api",
)

go_repository(
    name = "io_k8s_apiextensions_apiserver",
    commit = "80a4532647cb",
    importpath = "k8s.io/apiextensions-apiserver",
)

go_repository(
    name = "io_k8s_apimachinery",
    build_file_generation = "on",
    build_file_name = "BUILD.bazel",
    build_file_proto_mode = "disable_global",
    commit = "302974c03f7e",
    importpath = "k8s.io/apimachinery",
)

go_repository(
    name = "io_k8s_client_go",
    importpath = "k8s.io/client-go",
    tag = "v10.0.0",
)

go_repository(
    name = "io_k8s_klog",
    importpath = "k8s.io/klog",
    tag = "v0.1.0",
)

go_repository(
    name = "io_k8s_sigs_yaml",
    importpath = "sigs.k8s.io/yaml",
    tag = "v1.1.0",
)

go_repository(
    name = "io_k8s_test_infra",
    commit = "bfbc61258394",
    importpath = "k8s.io/test-infra",
)

go_repository(
    name = "io_opencensus_go",
    importpath = "go.opencensus.io",
    tag = "v0.13.0",
)

go_repository(
    name = "org_golang_google_api",
    commit = "ffa5046912fd",
    importpath = "google.golang.org/api",
)

go_repository(
    name = "org_golang_google_appengine",
    build_file_proto_mode = "disable_global",
    importpath = "google.golang.org/appengine",
    tag = "v1.0.0",
)

go_repository(
    name = "org_golang_google_genproto",
    commit = "11c7f9e547da",
    importpath = "google.golang.org/genproto",
)

go_repository(
    name = "org_golang_google_grpc",
    build_file_proto_mode = "disable_global",
    importpath = "google.golang.org/grpc",
    tag = "v1.7.2",
)

go_repository(
    name = "org_golang_x_crypto",
    commit = "cbcb75029529",
    importpath = "golang.org/x/crypto",
)

go_repository(
    name = "org_golang_x_net",
    commit = "afa5a82059c6",
    importpath = "golang.org/x/net",
)

go_repository(
    name = "org_golang_x_oauth2",
    commit = "9ff8ebcc8e24",
    importpath = "golang.org/x/oauth2",
)

go_repository(
    name = "org_golang_x_sync",
    commit = "112230192c58",
    importpath = "golang.org/x/sync",
)

go_repository(
    name = "org_golang_x_sys",
    commit = "953cdadca894",
    importpath = "golang.org/x/sys",
)

go_repository(
    name = "org_golang_x_text",
    importpath = "golang.org/x/text",
    tag = "v0.3.0",
)

go_repository(
    name = "org_golang_x_time",
    commit = "85acf8d2951c",
    importpath = "golang.org/x/time",
)

go_repository(
    name = "org_golang_x_tools",
    commit = "521d6ed310dd",
    importpath = "golang.org/x/tools",
)

go_repository(
    name = "cc_mvdan_interfacer",
    commit = "c20040233aed",
    importpath = "mvdan.cc/interfacer",
)

go_repository(
    name = "cc_mvdan_lint",
    commit = "adc824a0674b",
    importpath = "mvdan.cc/lint",
)

go_repository(
    name = "cc_mvdan_unparam",
    commit = "1b9ccfa71afe",
    importpath = "mvdan.cc/unparam",
)

go_repository(
    name = "co_honnef_go_tools",
    commit = "a1efa522b896",
    importpath = "honnef.co/go/tools",
)

go_repository(
    name = "com_4d63_gochecknoglobals",
    commit = "7c3491d2b6ec",
    importpath = "4d63.com/gochecknoglobals",
)

go_repository(
    name = "com_4d63_gochecknoinits",
    commit = "14d5915061e5",
    importpath = "4d63.com/gochecknoinits",
)

go_repository(
    name = "com_github_alecthomas_gocyclo",
    commit = "aa8f8b160214",
    importpath = "github.com/alecthomas/gocyclo",
)

go_repository(
    name = "com_github_alexflint_go_arg",
    importpath = "github.com/alexflint/go-arg",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_alexflint_go_scalar",
    importpath = "github.com/alexflint/go-scalar",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_alexkohler_nakedret",
    commit = "98ae56e4e0f3",
    importpath = "github.com/alexkohler/nakedret",
)

go_repository(
    name = "com_github_burntsushi_toml",
    importpath = "github.com/BurntSushi/toml",
    tag = "v0.3.1",
)

go_repository(
    name = "com_github_client9_misspell",
    importpath = "github.com/client9/misspell",
    tag = "v0.3.4",
)

go_repository(
    name = "com_github_gordonklaus_ineffassign",
    commit = "1003c8bd00dc",
    importpath = "github.com/gordonklaus/ineffassign",
)

go_repository(
    name = "com_github_jgautheron_goconst",
    commit = "9740945f5dcb",
    importpath = "github.com/jgautheron/goconst",
)

go_repository(
    name = "com_github_kisielk_errcheck",
    importpath = "github.com/kisielk/errcheck",
    tag = "v1.2.0",
)

go_repository(
    name = "com_github_kisielk_gotool",
    importpath = "github.com/kisielk/gotool",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_mdempsky_maligned",
    commit = "6e39bd26a8c8",
    importpath = "github.com/mdempsky/maligned",
)

go_repository(
    name = "com_github_mdempsky_unconvert",
    commit = "2f5dc3378ed3",
    importpath = "github.com/mdempsky/unconvert",
)

go_repository(
    name = "com_github_mibk_dupl",
    importpath = "github.com/mibk/dupl",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_mozilla_tls_observatory",
    commit = "a3c1b6cfecfd",
    importpath = "github.com/mozilla/tls-observatory",
)

go_repository(
    name = "com_github_nbutton23_zxcvbn_go",
    commit = "ae427f1e4c1d",
    importpath = "github.com/nbutton23/zxcvbn-go",
)

go_repository(
    name = "com_github_opennota_check",
    commit = "0c771f5545ff",
    importpath = "github.com/opennota/check",
)

go_repository(
    name = "com_github_rogpeppe_go_internal",
    importpath = "github.com/rogpeppe/go-internal",
    tag = "v1.3.0",
)

go_repository(
    name = "com_github_ryanuber_go_glob",
    importpath = "github.com/ryanuber/go-glob",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_securego_gosec",
    commit = "0ebfa2f8b7f8",
    importpath = "github.com/securego/gosec",
)

go_repository(
    name = "com_github_stripe_safesql",
    commit = "cddf355596fe",
    importpath = "github.com/stripe/safesql",
)

go_repository(
    name = "com_github_tsenart_deadcode",
    commit = "210d2dc333e9",
    importpath = "github.com/tsenart/deadcode",
)

go_repository(
    name = "com_github_walle_lll",
    commit = "8b13b3fbf731",
    importpath = "github.com/walle/lll",
)

go_repository(
    name = "in_gopkg_errgo_v2",
    importpath = "gopkg.in/errgo.v2",
    tag = "v2.1.0",
)

go_repository(
    name = "org_golang_x_lint",
    commit = "959b441ac422",
    importpath = "golang.org/x/lint",
)

go_repository(
    name = "com_github_google_renameio",
    importpath = "github.com/google/renameio",
    tag = "v0.1.0",
)

go_repository(
    name = "com_github_lib_pq",
    importpath = "github.com/lib/pq",
    tag = "v1.1.0",
)

go_repository(
    name = "org_golang_x_mod",
    commit = "4bf6d317e70e",
    importpath = "golang.org/x/mod",
)

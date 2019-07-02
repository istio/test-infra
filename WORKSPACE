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
    strip_prefix = "buildtools-0.25.1",
    url = "https://github.com/bazelbuild/buildtools/archive/0.25.1.zip",
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(go_version = "1.12.5")

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

load("@com_github_bazelbuild_buildtools//buildifier:deps.bzl", "buildifier_dependencies")

buildifier_dependencies()

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
    importpath = "github.com/beorn7/perks",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_bwmarrin_snowflake",
    importpath = "github.com/bwmarrin/snowflake",
    tag = "v0.0.0",
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
    tag = "v1.2.1",
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
    tag = "v0.3.0",
)

go_repository(
    name = "com_github_google_go_github",
    importpath = "github.com/google/go-github",
    tag = "v17.0.0",
)

go_repository(
    name = "com_github_google_go_querystring",
    importpath = "github.com/google/go-querystring",
    tag = "v1.0.0",
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
    importpath = "github.com/gorilla/context",
    tag = "v1.1.1",
)

go_repository(
    name = "com_github_gorilla_securecookie",
    importpath = "github.com/gorilla/securecookie",
    tag = "v1.1.1",
)

go_repository(
    name = "com_github_gorilla_sessions",
    importpath = "github.com/gorilla/sessions",
    tag = "v1.1.3",
)

go_repository(
    name = "com_github_gregjones_httpcache",
    commit = "3befbb6ad0cc",
    importpath = "github.com/gregjones/httpcache",
)

go_repository(
    name = "com_github_hashicorp_errwrap",
    commit = "7554cd9344ce",
    importpath = "github.com/hashicorp/errwrap",
)

go_repository(
    name = "com_github_hashicorp_go_multierror",
    commit = "b7773ae21874",
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
    tag = "v1.1.6",
)

go_repository(
    name = "com_github_knative_build",
    commit = "38ace00371c7",
    importpath = "github.com/knative/build",
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
    importpath = "github.com/mattn/go-zglob",
    tag = "v0.0.1",
)

go_repository(
    name = "com_github_matttproud_golang_protobuf_extensions",
    importpath = "github.com/matttproud/golang_protobuf_extensions",
    tag = "v1.0.1",
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
    tag = "v1.7.0",
)

go_repository(
    name = "com_github_onsi_gomega",
    importpath = "github.com/onsi/gomega",
    tag = "v1.4.3",
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
    tag = "v0.9.4",
)

go_repository(
    name = "com_github_prometheus_client_model",
    commit = "fd36f4220a90",
    importpath = "github.com/prometheus/client_model",
)

go_repository(
    name = "com_github_prometheus_common",
    importpath = "github.com/prometheus/common",
    tag = "v0.4.1",
)

go_repository(
    name = "com_github_prometheus_procfs",
    importpath = "github.com/prometheus/procfs",
    tag = "v0.0.3",
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
    commit = "9e1955d9fb6e",
    importpath = "github.com/shurcooL/go",
)

go_repository(
    name = "com_github_shurcool_graphql",
    commit = "e4a3a37e6d42",
    importpath = "github.com/shurcooL/graphql",
)

go_repository(
    name = "com_github_sirupsen_logrus",
    importpath = "github.com/sirupsen/logrus",
    tag = "v1.4.2",
)

go_repository(
    name = "com_github_spf13_pflag",
    importpath = "github.com/spf13/pflag",
    tag = "v1.0.3",
)

go_repository(
    name = "com_github_stretchr_objx",
    importpath = "github.com/stretchr/objx",
    tag = "v0.1.1",
)

go_repository(
    name = "com_github_stretchr_testify",
    importpath = "github.com/stretchr/testify",
    tag = "v1.3.0",
)

go_repository(
    name = "com_google_cloud_go",
    importpath = "cloud.google.com/go",
    tag = "v0.37.4",
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
    commit = "6db15a15d2d3",
    importpath = "k8s.io/api",
)

go_repository(
    name = "io_k8s_apiextensions_apiserver",
    commit = "1f84094d7e8e",
    importpath = "k8s.io/apiextensions-apiserver",
)

go_repository(
    name = "io_k8s_apimachinery",
    build_file_generation = "on",
    build_file_name = "BUILD.bazel",
    build_file_proto_mode = "disable_global",
    commit = "49ce2735e507",
    importpath = "k8s.io/apimachinery",
)

go_repository(
    name = "io_k8s_client_go",
    importpath = "k8s.io/client-go",
    tag = "v9.0.0",
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
    commit = "68bb362e4b66",
    importpath = "k8s.io/test-infra",
)

go_repository(
    name = "io_opencensus_go",
    importpath = "go.opencensus.io",
    tag = "v0.20.2",
)

go_repository(
    name = "org_golang_google_api",
    importpath = "google.golang.org/api",
    tag = "v0.3.2",
)

go_repository(
    name = "org_golang_google_appengine",
    build_file_proto_mode = "disable_global",
    importpath = "google.golang.org/appengine",
    tag = "v1.4.0",
)

go_repository(
    name = "org_golang_google_genproto",
    commit = "64821d5d2107",
    importpath = "google.golang.org/genproto",
)

go_repository(
    name = "org_golang_google_grpc",
    build_file_proto_mode = "disable_global",
    importpath = "google.golang.org/grpc",
    tag = "v1.19.1",
)

go_repository(
    name = "org_golang_x_crypto",
    commit = "38d8ce5564a5",
    importpath = "golang.org/x/crypto",
)

go_repository(
    name = "org_golang_x_net",
    commit = "b630fd6fe46b",
    importpath = "golang.org/x/net",
)

go_repository(
    name = "org_golang_x_oauth2",
    commit = "e64efc72b421",
    importpath = "golang.org/x/oauth2",
)

go_repository(
    name = "org_golang_x_sync",
    commit = "e225da77a7e6",
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
    tag = "v0.3.2",
)

go_repository(
    name = "org_golang_x_time",
    commit = "85acf8d2951c",
    importpath = "golang.org/x/time",
)

go_repository(
    name = "org_golang_x_tools",
    commit = "923d25813098",
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
    commit = "cc9d2fb52971",
    importpath = "mvdan.cc/unparam",
)

go_repository(
    name = "co_honnef_go_tools",
    commit = "3f1c8253044a",
    importpath = "honnef.co/go/tools",
)

go_repository(
    name = "com_4d63_gochecknoglobals",
    commit = "abbdf6ec0afb",
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
    commit = "c0e305a4f690",
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
    tag = "v1.1.0",
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
    commit = "2db5a8ead8e7",
    importpath = "github.com/mdempsky/unconvert",
)

go_repository(
    name = "com_github_mibk_dupl",
    importpath = "github.com/mibk/dupl",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_mozilla_tls_observatory",
    commit = "8791a200eb40",
    importpath = "github.com/mozilla/tls-observatory",
)

go_repository(
    name = "com_github_nbutton23_zxcvbn_go",
    commit = "a22cb81b2ecd",
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
    tag = "v1.1.0",
)

go_repository(
    name = "com_github_ryanuber_go_glob",
    commit = "256dc444b735",
    importpath = "github.com/ryanuber/go-glob",
)

go_repository(
    name = "com_github_securego_gosec",
    commit = "a966ff760c3a",
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
    commit = "5614ed5bae6f",
    importpath = "golang.org/x/lint",
)

go_repository(
    name = "cc_mvdan_xurls_v2",
    importpath = "mvdan.cc/xurls/v2",
    tag = "v2.0.0",
)

go_repository(
    name = "com_github_alecthomas_template",
    commit = "a0175ee3bccc",
    importpath = "github.com/alecthomas/template",
)

go_repository(
    name = "com_github_alecthomas_units",
    commit = "2efee857e7cf",
    importpath = "github.com/alecthomas/units",
)

go_repository(
    name = "com_github_andygrunwald_go_gerrit",
    commit = "174420ebee6c",
    importpath = "github.com/andygrunwald/go-gerrit",
)

go_repository(
    name = "com_github_apache_thrift",
    importpath = "github.com/apache/thrift",
    tag = "v0.12.0",
)

go_repository(
    name = "com_github_armon_consul_api",
    commit = "eb2c6b5be1b6",
    importpath = "github.com/armon/consul-api",
)

go_repository(
    name = "com_github_aws_aws_k8s_tester",
    commit = "b411acf57dfe",
    importpath = "github.com/aws/aws-k8s-tester",
)

go_repository(
    name = "com_github_aws_aws_sdk_go",
    importpath = "github.com/aws/aws-sdk-go",
    tag = "v1.16.36",
)

go_repository(
    name = "com_github_azure_azure_pipeline_go",
    importpath = "github.com/Azure/azure-pipeline-go",
    tag = "v0.1.9",
)

go_repository(
    name = "com_github_azure_azure_sdk_for_go",
    importpath = "github.com/Azure/azure-sdk-for-go",
    tag = "v21.1.0",
)

go_repository(
    name = "com_github_azure_azure_storage_blob_go",
    commit = "457680cc0804",
    importpath = "github.com/Azure/azure-storage-blob-go",
)

go_repository(
    name = "com_github_azure_go_autorest",
    importpath = "github.com/Azure/go-autorest",
    tag = "v10.15.5",
)

go_repository(
    name = "com_github_bazelbuild_bazel_gazelle",
    commit = "e530fae7ce5c",
    importpath = "github.com/bazelbuild/bazel-gazelle",
)

go_repository(
    name = "com_github_bgentry_speakeasy",
    importpath = "github.com/bgentry/speakeasy",
    tag = "v0.1.0",
)

go_repository(
    name = "com_github_blang_semver",
    importpath = "github.com/blang/semver",
    tag = "v3.5.1",
)

go_repository(
    name = "com_github_coreos_etcd",
    importpath = "github.com/coreos/etcd",
    tag = "v3.3.10",
)

go_repository(
    name = "com_github_coreos_go_etcd",
    importpath = "github.com/coreos/go-etcd",
    tag = "v2.0.0",
)

go_repository(
    name = "com_github_coreos_go_semver",
    importpath = "github.com/coreos/go-semver",
    tag = "v0.2.0",
)

go_repository(
    name = "com_github_coreos_go_systemd",
    commit = "39ca1b05acc7",
    importpath = "github.com/coreos/go-systemd",
)

go_repository(
    name = "com_github_coreos_pkg",
    commit = "3ac0863d7acf",
    importpath = "github.com/coreos/pkg",
)

go_repository(
    name = "com_github_cpuguy83_go_md2man",
    importpath = "github.com/cpuguy83/go-md2man",
    tag = "v1.0.10",
)

go_repository(
    name = "com_github_denisenkom_go_mssqldb",
    commit = "2fea367d496d",
    importpath = "github.com/denisenkom/go-mssqldb",
)

go_repository(
    name = "com_github_dgrijalva_jwt_go",
    importpath = "github.com/dgrijalva/jwt-go",
    tag = "v3.2.0",
)

go_repository(
    name = "com_github_djherbis_atime",
    importpath = "github.com/djherbis/atime",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_docker_distribution",
    commit = "edc3ab29cdff",
    importpath = "github.com/docker/distribution",
)

go_repository(
    name = "com_github_docker_docker",
    commit = "5e5fadb3c020",
    importpath = "github.com/docker/docker",
)

go_repository(
    name = "com_github_docker_go_connections",
    importpath = "github.com/docker/go-connections",
    tag = "v0.3.0",
)

go_repository(
    name = "com_github_docker_go_units",
    importpath = "github.com/docker/go-units",
    tag = "v0.3.3",
)

go_repository(
    name = "com_github_dustin_go_humanize",
    importpath = "github.com/dustin/go-humanize",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_eapache_go_resiliency",
    importpath = "github.com/eapache/go-resiliency",
    tag = "v1.1.0",
)

go_repository(
    name = "com_github_eapache_go_xerial_snappy",
    commit = "776d5712da21",
    importpath = "github.com/eapache/go-xerial-snappy",
)

go_repository(
    name = "com_github_eapache_queue",
    importpath = "github.com/eapache/queue",
    tag = "v1.1.0",
)

go_repository(
    name = "com_github_erikstmartin_go_testdb",
    commit = "8d10e4a1bae5",
    importpath = "github.com/erikstmartin/go-testdb",
)

go_repository(
    name = "com_github_evanphx_json_patch",
    importpath = "github.com/evanphx/json-patch",
    tag = "v4.2.0",
)

go_repository(
    name = "com_github_fatih_color",
    importpath = "github.com/fatih/color",
    tag = "v1.7.0",
)

go_repository(
    name = "com_github_fsouza_fake_gcs_server",
    commit = "e85be23bdaa8",
    importpath = "github.com/fsouza/fake-gcs-server",
)

go_repository(
    name = "com_github_go_kit_kit",
    importpath = "github.com/go-kit/kit",
    tag = "v0.8.0",
)

go_repository(
    name = "com_github_go_logfmt_logfmt",
    importpath = "github.com/go-logfmt/logfmt",
    tag = "v0.3.0",
)

go_repository(
    name = "com_github_go_openapi_jsonpointer",
    commit = "779f45308c19",
    importpath = "github.com/go-openapi/jsonpointer",
)

go_repository(
    name = "com_github_go_openapi_jsonreference",
    commit = "36d33bfe519e",
    importpath = "github.com/go-openapi/jsonreference",
)

go_repository(
    name = "com_github_go_openapi_spec",
    commit = "fa03337d7da5",
    importpath = "github.com/go-openapi/spec",
)

go_repository(
    name = "com_github_go_openapi_swag",
    commit = "cf0bdb963811",
    importpath = "github.com/go-openapi/swag",
)

go_repository(
    name = "com_github_go_sql_driver_mysql",
    commit = "7ebe0a500653",
    importpath = "github.com/go-sql-driver/mysql",
)

go_repository(
    name = "com_github_go_stack_stack",
    importpath = "github.com/go-stack/stack",
    tag = "v1.8.0",
)

go_repository(
    name = "com_github_gobuffalo_envy",
    importpath = "github.com/gobuffalo/envy",
    tag = "v1.6.15",
)

go_repository(
    name = "com_github_golang_groupcache",
    commit = "24b0969c4cb7",
    importpath = "github.com/golang/groupcache",
)

go_repository(
    name = "com_github_golang_mock",
    importpath = "github.com/golang/mock",
    tag = "v1.2.0",
)

go_repository(
    name = "com_github_golang_snappy",
    commit = "2e65f85255db",
    importpath = "github.com/golang/snappy",
)

go_repository(
    name = "com_github_gomodule_redigo",
    importpath = "github.com/gomodule/redigo",
    tag = "v1.7.0",
)

go_repository(
    name = "com_github_google_go_containerregistry",
    commit = "f1df91a4a813",
    importpath = "github.com/google/go-containerregistry",
)

go_repository(
    name = "com_github_google_martian",
    importpath = "github.com/google/martian",
    tag = "v2.1.0",
)

go_repository(
    name = "com_github_google_pprof",
    commit = "3ea8567a2e57",
    importpath = "github.com/google/pprof",
)

go_repository(
    name = "com_github_google_uuid",
    importpath = "github.com/google/uuid",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_googleapis_gax_go_v2",
    importpath = "github.com/googleapis/gax-go/v2",
    tag = "v2.0.5",
)

go_repository(
    name = "com_github_gophercloud_gophercloud",
    commit = "bdd8b1ecd793",
    importpath = "github.com/gophercloud/gophercloud",
)

go_repository(
    name = "com_github_gorilla_mux",
    importpath = "github.com/gorilla/mux",
    tag = "v1.6.2",
)

go_repository(
    name = "com_github_gorilla_websocket",
    commit = "4201258b820c",
    importpath = "github.com/gorilla/websocket",
)

go_repository(
    name = "com_github_grpc_ecosystem_go_grpc_middleware",
    importpath = "github.com/grpc-ecosystem/go-grpc-middleware",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_grpc_ecosystem_go_grpc_prometheus",
    importpath = "github.com/grpc-ecosystem/go-grpc-prometheus",
    tag = "v1.2.0",
)

go_repository(
    name = "com_github_grpc_ecosystem_grpc_gateway",
    importpath = "github.com/grpc-ecosystem/grpc-gateway",
    tag = "v1.4.1",
)

go_repository(
    name = "com_github_hashicorp_golang_lru",
    importpath = "github.com/hashicorp/golang-lru",
    tag = "v0.5.1",
)

go_repository(
    name = "com_github_hashicorp_hcl",
    importpath = "github.com/hashicorp/hcl",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_inconshreveable_mousetrap",
    importpath = "github.com/inconshreveable/mousetrap",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_influxdata_influxdb",
    commit = "049f9b42e9a5",
    importpath = "github.com/influxdata/influxdb",
)

go_repository(
    name = "com_github_jinzhu_gorm",
    commit = "572d0a0ab1eb",
    importpath = "github.com/jinzhu/gorm",
)

go_repository(
    name = "com_github_jinzhu_inflection",
    commit = "f5c5f50e6090",
    importpath = "github.com/jinzhu/inflection",
)

go_repository(
    name = "com_github_jinzhu_now",
    importpath = "github.com/jinzhu/now",
    tag = "v1.0.1",
)

go_repository(
    name = "com_github_jmespath_go_jmespath",
    commit = "c2b33e8439af",
    importpath = "github.com/jmespath/go-jmespath",
)

go_repository(
    name = "com_github_joho_godotenv",
    importpath = "github.com/joho/godotenv",
    tag = "v1.3.0",
)

go_repository(
    name = "com_github_jonboulle_clockwork",
    importpath = "github.com/jonboulle/clockwork",
    tag = "v0.1.0",
)

go_repository(
    name = "com_github_jstemmer_go_junit_report",
    commit = "af01ea7f8024",
    importpath = "github.com/jstemmer/go-junit-report",
)

go_repository(
    name = "com_github_julienschmidt_httprouter",
    importpath = "github.com/julienschmidt/httprouter",
    tag = "v1.2.0",
)

go_repository(
    name = "com_github_klauspost_compress",
    importpath = "github.com/klauspost/compress",
    tag = "v1.4.1",
)

go_repository(
    name = "com_github_klauspost_cpuid",
    importpath = "github.com/klauspost/cpuid",
    tag = "v1.2.1",
)

go_repository(
    name = "com_github_klauspost_pgzip",
    importpath = "github.com/klauspost/pgzip",
    tag = "v1.2.1",
)

go_repository(
    name = "com_github_knative_pkg",
    commit = "916205998db9",
    importpath = "github.com/knative/pkg",
)

go_repository(
    name = "com_github_konsorten_go_windows_terminal_sequences",
    importpath = "github.com/konsorten/go-windows-terminal-sequences",
    tag = "v1.0.2",
)

go_repository(
    name = "com_github_kr_logfmt",
    commit = "b84e30acd515",
    importpath = "github.com/kr/logfmt",
)

go_repository(
    name = "com_github_lib_pq",
    importpath = "github.com/lib/pq",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_magiconair_properties",
    importpath = "github.com/magiconair/properties",
    tag = "v1.8.0",
)

go_repository(
    name = "com_github_mailru_easyjson",
    commit = "32fa128f234d",
    importpath = "github.com/mailru/easyjson",
)

go_repository(
    name = "com_github_markbates_inflect",
    importpath = "github.com/markbates/inflect",
    tag = "v1.0.4",
)

go_repository(
    name = "com_github_mattbaird_jsonpatch",
    commit = "81af80346b1a",
    importpath = "github.com/mattbaird/jsonpatch",
)

go_repository(
    name = "com_github_mattn_go_colorable",
    importpath = "github.com/mattn/go-colorable",
    tag = "v0.0.9",
)

go_repository(
    name = "com_github_mattn_go_isatty",
    importpath = "github.com/mattn/go-isatty",
    tag = "v0.0.4",
)

go_repository(
    name = "com_github_mattn_go_runewidth",
    importpath = "github.com/mattn/go-runewidth",
    tag = "v0.0.2",
)

go_repository(
    name = "com_github_mattn_go_sqlite3",
    commit = "38ee283dabf1",
    importpath = "github.com/mattn/go-sqlite3",
)

go_repository(
    name = "com_github_microsoft_go_winio",
    importpath = "github.com/Microsoft/go-winio",
    tag = "v0.4.12",
)

go_repository(
    name = "com_github_mitchellh_go_homedir",
    importpath = "github.com/mitchellh/go-homedir",
    tag = "v1.1.0",
)

go_repository(
    name = "com_github_mitchellh_ioprogress",
    commit = "6a23b12fa88e",
    importpath = "github.com/mitchellh/ioprogress",
)

go_repository(
    name = "com_github_mitchellh_mapstructure",
    importpath = "github.com/mitchellh/mapstructure",
    tag = "v1.1.2",
)

go_repository(
    name = "com_github_mwitkow_go_conntrack",
    commit = "cc309e4a2223",
    importpath = "github.com/mwitkow/go-conntrack",
)

go_repository(
    name = "com_github_nytimes_gziphandler",
    commit = "63027b26b87e",
    importpath = "github.com/NYTimes/gziphandler",
)

go_repository(
    name = "com_github_olekukonko_tablewriter",
    commit = "a0225b3f23b5",
    importpath = "github.com/olekukonko/tablewriter",
)

go_repository(
    name = "com_github_opencontainers_go_digest",
    importpath = "github.com/opencontainers/go-digest",
    tag = "v1.0.0-rc1",
)

go_repository(
    name = "com_github_opencontainers_image_spec",
    importpath = "github.com/opencontainers/image-spec",
    tag = "v1.0.1",
)

go_repository(
    name = "com_github_openzipkin_zipkin_go",
    importpath = "github.com/openzipkin/zipkin-go",
    tag = "v0.1.6",
)

go_repository(
    name = "com_github_pelletier_go_toml",
    importpath = "github.com/pelletier/go-toml",
    tag = "v1.3.0",
)

go_repository(
    name = "com_github_pierrec_lz4",
    importpath = "github.com/pierrec/lz4",
    tag = "v2.0.5",
)

go_repository(
    name = "com_github_pkg_errors",
    importpath = "github.com/pkg/errors",
    tag = "v0.8.1",
)

go_repository(
    name = "com_github_puerkitobio_purell",
    importpath = "github.com/PuerkitoBio/purell",
    tag = "v1.1.1",
)

go_repository(
    name = "com_github_puerkitobio_urlesc",
    commit = "de5bf2ad4578",
    importpath = "github.com/PuerkitoBio/urlesc",
)

go_repository(
    name = "com_github_rcrowley_go_metrics",
    commit = "3113b8401b8a",
    importpath = "github.com/rcrowley/go-metrics",
)

go_repository(
    name = "com_github_russross_blackfriday",
    importpath = "github.com/russross/blackfriday",
    tag = "v1.5.2",
)

go_repository(
    name = "com_github_shopify_sarama",
    importpath = "github.com/Shopify/sarama",
    tag = "v1.19.0",
)

go_repository(
    name = "com_github_shopify_toxiproxy",
    importpath = "github.com/Shopify/toxiproxy",
    tag = "v2.1.4",
)

go_repository(
    name = "com_github_soheilhy_cmux",
    importpath = "github.com/soheilhy/cmux",
    tag = "v0.1.4",
)

go_repository(
    name = "com_github_spf13_afero",
    importpath = "github.com/spf13/afero",
    tag = "v1.1.2",
)

go_repository(
    name = "com_github_spf13_cast",
    importpath = "github.com/spf13/cast",
    tag = "v1.3.0",
)

go_repository(
    name = "com_github_spf13_cobra",
    importpath = "github.com/spf13/cobra",
    tag = "v0.0.5",
)

go_repository(
    name = "com_github_spf13_jwalterweatherman",
    importpath = "github.com/spf13/jwalterweatherman",
    tag = "v1.0.0",
)

go_repository(
    name = "com_github_spf13_viper",
    importpath = "github.com/spf13/viper",
    tag = "v1.3.2",
)

go_repository(
    name = "com_github_tektoncd_pipeline",
    commit = "7c43fbae2816",
    importpath = "github.com/tektoncd/pipeline",
)

go_repository(
    name = "com_github_tmc_grpc_websocket_proxy",
    commit = "89b8d40f7ca8",
    importpath = "github.com/tmc/grpc-websocket-proxy",
)

go_repository(
    name = "com_github_ugorji_go",
    importpath = "github.com/ugorji/go",
    tag = "v1.1.1",
)

go_repository(
    name = "com_github_ugorji_go_codec",
    commit = "d75b2dcb6bc8",
    importpath = "github.com/ugorji/go/codec",
)

go_repository(
    name = "com_github_urfave_cli",
    importpath = "github.com/urfave/cli",
    tag = "v1.18.0",
)

go_repository(
    name = "com_github_xiang90_probing",
    commit = "07dd2e8dfe18",
    importpath = "github.com/xiang90/probing",
)

go_repository(
    name = "com_github_xlab_handysort",
    commit = "fb3537ed64a1",
    importpath = "github.com/xlab/handysort",
)

go_repository(
    name = "com_github_xordataexchange_crypt",
    commit = "b2862e3d0a77",
    importpath = "github.com/xordataexchange/crypt",
)

go_repository(
    name = "in_gopkg_alecthomas_kingpin_v2",
    importpath = "gopkg.in/alecthomas/kingpin.v2",
    tag = "v2.2.6",
)

go_repository(
    name = "in_gopkg_cheggaaa_pb_v1",
    importpath = "gopkg.in/cheggaaa/pb.v1",
    tag = "v1.0.25",
)

go_repository(
    name = "io_etcd_go_bbolt",
    importpath = "go.etcd.io/bbolt",
    tag = "v1.3.1-etcd.7",
)

go_repository(
    name = "io_etcd_go_etcd",
    commit = "83304cfc808c",
    importpath = "go.etcd.io/etcd",
)

go_repository(
    name = "io_k8s_code_generator",
    commit = "b1289fc74931",
    importpath = "k8s.io/code-generator",
)

go_repository(
    name = "io_k8s_gengo",
    commit = "7a1b7fb0289f",
    importpath = "k8s.io/gengo",
)

go_repository(
    name = "io_k8s_kube_openapi",
    commit = "0cf8f7e6ed1d",
    importpath = "k8s.io/kube-openapi",
)

go_repository(
    name = "io_k8s_kubernetes",
    importpath = "k8s.io/kubernetes",
    tag = "v1.14.3",
)

go_repository(
    name = "io_k8s_repo_infra",
    commit = "df02ded38f95",
    importpath = "k8s.io/repo-infra",
)

go_repository(
    name = "io_k8s_utils",
    commit = "5e321f9a457c",
    importpath = "k8s.io/utils",
)

go_repository(
    name = "ml_vbom_util",
    commit = "256737ac55c4",
    importpath = "vbom.ml/util",
)

go_repository(
    name = "org_golang_x_exp",
    commit = "509febef88a4",
    importpath = "golang.org/x/exp",
)

go_repository(
    name = "org_uber_go_atomic",
    importpath = "go.uber.org/atomic",
    tag = "v1.3.2",
)

go_repository(
    name = "org_uber_go_multierr",
    importpath = "go.uber.org/multierr",
    tag = "v1.1.0",
)

go_repository(
    name = "org_uber_go_zap",
    importpath = "go.uber.org/zap",
    tag = "v1.9.1",
)

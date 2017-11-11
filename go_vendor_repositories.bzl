load("@io_bazel_rules_go//go:def.bzl", "go_repository")

def go_vendor_repositories():
  go_repository(
    name = "com_google_cloud_go",
    commit = "2d3a6656c17a60b0815b7e06ab0be04eacb6e613",
    importpath = "cloud.google.com/go",
  )

  go_repository(
    name = "com_github_PuerkitoBio_purell",
    commit = "0bcb03f4b4d0a9428594752bd2a3b9aa0a9d4bd4",
    importpath = "github.com/PuerkitoBio/purell",
  )

  go_repository(
    name = "com_github_PuerkitoBio_urlesc",
    commit = "de5bf2ad457846296e2031421a34e2568e304e35",
    importpath = "github.com/PuerkitoBio/urlesc",
  )

  go_repository(
    name = "com_github_beorn7_perks",
    commit = "4c0e84591b9aa9e6dcfdf3e020114cd81f89d5f9",
    importpath = "github.com/beorn7/perks",
  )

  go_repository(
    name = "com_github_davecgh_go_spew",
    commit = "346938d642f2ec3594ed81d874461961cd0faa76",
    importpath = "github.com/davecgh/go-spew",
  )

  go_repository(
    name = "com_github_emicklei_go_restful",
    commit = "5741799b275a3c4a5a9623a993576d7545cf7b5c",
    importpath = "github.com/emicklei/go-restful",
  )

  go_repository(
    name = "com_github_emicklei_go_restful_swagger12",
    commit = "dcef7f55730566d41eae5db10e7d6981829720f6",
    importpath = "github.com/emicklei/go-restful-swagger12",
  )

  go_repository(
    name = "com_github_ghodss_yaml",
    commit = "0ca9ea5df5451ffdf184b4428c902747c2c11cd7",
    importpath = "github.com/ghodss/yaml",
  )

  go_repository(
    name = "com_github_go_openapi_jsonpointer",
    commit = "779f45308c19820f1a69e9a4cd965f496e0da10f",
    importpath = "github.com/go-openapi/jsonpointer",
  )

  go_repository(
    name = "com_github_go_openapi_jsonreference",
    commit = "36d33bfe519efae5632669801b180bf1a245da3b",
    importpath = "github.com/go-openapi/jsonreference",
  )

  go_repository(
    name = "com_github_go_openapi_spec",
    commit = "a4fa9574c7aa73b2fc54e251eb9524d0482bb592",
    importpath = "github.com/go-openapi/spec",
  )

  go_repository(
    name = "com_github_go_openapi_swag",
    commit = "f3f9494671f93fcff853e3c6e9e948b3eb71e590",
    importpath = "github.com/go-openapi/swag",
  )

  go_repository(
    name = "com_github_gogo_protobuf",
    commit = "342cbe0a04158f6dcb03ca0079991a51a4248c02",
    importpath = "github.com/gogo/protobuf",
  )

  go_repository(
    name = "com_github_golang_glog",
    commit = "23def4e6c14b4da8ac2ed8007337bc5eb5007998",
    importpath = "github.com/golang/glog",
  )

  go_repository(
    name = "com_github_golang_protobuf",
    commit = "1643683e1b54a9e88ad26d98f81400c8c9d9f4f9",
    importpath = "github.com/golang/protobuf",
  )

  go_repository(
    name = "com_github_google_btree",
    commit = "316fb6d3f031ae8f4d457c6c5186b9e3ded70435",
    importpath = "github.com/google/btree",
  )

  go_repository(
    name = "com_github_google_go_github",
    commit = "8c08f4fba5e05e0fd2821a5f80cf0cf643bd5314",
    importpath = "github.com/google/go-github",
  )

  go_repository(
    name = "com_github_google_go_querystring",
    commit = "53e6ce116135b80d037921a7fdd5138cf32d7a8a",
    importpath = "github.com/google/go-querystring",
  )

  go_repository(
    name = "com_github_google_gofuzz",
    commit = "24818f796faf91cd76ec7bddd72458fbced7a6c1",
    importpath = "github.com/google/gofuzz",
  )

  go_repository(
    name = "com_github_googleapis_gax_go",
    commit = "317e0006254c44a0ac427cc52a0e083ff0b9622f",
    importpath = "github.com/googleapis/gax-go",
  )

  go_repository(
    name = "com_github_googleapis_gnostic",
    commit = "ee43cbb60db7bd22502942cccbc39059117352ab",
    importpath = "github.com/googleapis/gnostic",
  )

  go_repository(
    name = "com_github_gregjones_httpcache",
    commit = "22a0b1feae53974ed4cfe27bcce70dba061cc5fd",
    importpath = "github.com/gregjones/httpcache",
  )

  go_repository(
    name = "com_github_hashicorp_errwrap",
    commit = "7554cd9344cec97297fa6649b055a8c98c2a1e55",
    importpath = "github.com/hashicorp/errwrap",
  )

  go_repository(
    name = "com_github_hashicorp_go_multierror",
    commit = "83588e72410abfbe4df460eeb6f30841ae47d4c4",
    importpath = "github.com/hashicorp/go-multierror",
  )

  go_repository(
    name = "com_github_hashicorp_golang_lru",
    commit = "0a025b7e63adc15a622f29b0b2c4c3848243bbf6",
    importpath = "github.com/hashicorp/golang-lru",
  )

  go_repository(
    name = "com_github_howeyc_gopass",
    commit = "bf9dde6d0d2c004a008c27aaee91170c786f6db8",
    importpath = "github.com/howeyc/gopass",
  )

  go_repository(
    name = "com_github_imdario_mergo",
    commit = "7fe0c75c13abdee74b09fcacef5ea1c6bba6a874",
    importpath = "github.com/imdario/mergo",
  )

  go_repository(
    name = "com_github_json_iterator_go",
    commit = "6240e1e7983a85228f7fd9c3e1b6932d46ec58e2",
    importpath = "github.com/json-iterator/go",
  )

  go_repository(
    name = "com_github_juju_ratelimit",
    commit = "59fac5042749a5afb9af70e813da1dd5474f0167",
    importpath = "github.com/juju/ratelimit",
  )

  go_repository(
    name = "com_github_mailru_easyjson",
    commit = "5f62e4f3afa2f576dc86531b7df4d966b19ef8f8",
    importpath = "github.com/mailru/easyjson",
  )

  go_repository(
    name = "com_github_matttproud_golang_protobuf_extensions",
    commit = "3247c84500bff8d9fb6d579d800f20b3e091582c",
    importpath = "github.com/matttproud/golang_protobuf_extensions",
  )

  go_repository(
    name = "com_github_petar_GoLLRB",
    commit = "53be0d36a84c2a886ca057d34b6aa4468df9ccb4",
    importpath = "github.com/petar/GoLLRB",
  )

  go_repository(
    name = "com_github_peterbourgon_diskv",
    commit = "5f041e8faa004a95c88a202771f4cc3e991971e6",
    importpath = "github.com/peterbourgon/diskv",
  )

  go_repository(
    name = "com_github_prometheus_client_golang",
    commit = "c5b7fccd204277076155f10851dad72b76a49317",
    importpath = "github.com/prometheus/client_golang",
  )

  go_repository(
    name = "com_github_prometheus_client_model",
    commit = "6f3806018612930941127f2a7c6c453ba2c527d2",
    importpath = "github.com/prometheus/client_model",
  )

  go_repository(
    name = "com_github_prometheus_common",
    commit = "e3fb1a1acd7605367a2b378bc2e2f893c05174b7",
    importpath = "github.com/prometheus/common",
  )

  go_repository(
    name = "com_github_prometheus_procfs",
    commit = "a6e9df898b1336106c743392c48ee0b71f5c4efa",
    importpath = "github.com/prometheus/procfs",
  )

  go_repository(
    name = "com_github_spf13_pflag",
    commit = "e57e3eeb33f795204c1ca35f56c44f83227c6e66",
    importpath = "github.com/spf13/pflag",
  )

  go_repository(
    name = "org_golang_x_crypto",
    commit = "6a293f2d4b14b8e6d3f0539e383f6d0d30fce3fd",
    importpath = "golang.org/x/crypto",
  )

  go_repository(
    name = "org_golang_x_net",
    commit = "a337091b0525af65de94df2eb7e98bd9962dcbe2",
    importpath = "golang.org/x/net",
  )

  go_repository(
    name = "org_golang_x_oauth2",
    commit = "9ff8ebcc8e241d46f52ecc5bff0e5a2f2dbef402",
    importpath = "golang.org/x/oauth2",
  )

  go_repository(
    name = "org_golang_x_sys",
    commit = "1e2299c37cc91a509f1b12369872d27be0ce98a6",
    importpath = "golang.org/x/sys",
  )

  go_repository(
    name = "org_golang_x_text",
    commit = "88f656faf3f37f690df1a32515b479415e1a6769",
    importpath = "golang.org/x/text",
  )

  go_repository(
    name = "org_golang_google_api",
    commit = "3b6ce7577f7305c6ba51dce053082c2aed563378",
    importpath = "google.golang.org/api",
  )

  go_repository(
    name = "org_golang_google_appengine",
    commit = "150dc57a1b433e64154302bdc40b6bb8aefa313a",
    importpath = "google.golang.org/appengine",
  )

  go_repository(
    name = "org_golang_google_genproto",
    commit = "11c7f9e547da6db876260ce49ea7536985904c9b",
    importpath = "google.golang.org/genproto",
  )

  go_repository(
    name = "org_golang_google_grpc",
    commit = "5ffe3083946d5603a0578721101dc8165b1d5b5f",
    importpath = "google.golang.org/grpc",
  )

  go_repository(
    name = "in_gopkg_inf_v0",
    commit = "3887ee99ecf07df5b447e9b00d9c0b2adaa9f3e4",
    importpath = "gopkg.in/inf.v0",
  )

  go_repository(
    name = "in_gopkg_yaml_v2",
    commit = "eb3733d160e74a9c7e442f435eb3bea458e1d19f",
    importpath = "gopkg.in/yaml.v2",
  )

  go_repository(
    name = "io_k8s_api",
    commit = "218912509d74a117d05a718bb926d0948e531c20",
    importpath = "k8s.io/api",
  )

  go_repository(
    name = "io_k8s_apiextensions_apiserver",
    commit = "51a1910459f074162eb4e25233e461fe91e99405",
    importpath = "k8s.io/apiextensions-apiserver",
  )

  go_repository(
    name = "io_k8s_apimachinery",
    commit = "18a564baac720819100827c16fdebcadb05b2d0d",
    importpath = "k8s.io/apimachinery",
  )

  go_repository(
    name = "io_k8s_client_go",
    commit = "2ae454230481a7cb5544325e12ad7658ecccd19b",
    importpath = "k8s.io/client-go",
  )

  go_repository(
    name = "io_k8s_kube_openapi",
    commit = "39a7bf85c140f972372c2a0d1ee40adbf0c8bfe1",
    importpath = "k8s.io/kube-openapi",
  )

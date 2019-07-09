load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_repositories():
    go_repository(
        name = "com_github_beorn7_perks",
        importpath = "github.com/beorn7/perks",
        sum = "h1:BtpsbiV638WQZwhA98cEZw2BsbnQJrbd0BI7tsy0W1c=",
        version = "v0.0.0-20160804104726-4c0e84591b9a",
    )

    go_repository(
        name = "com_github_bwmarrin_snowflake",
        importpath = "github.com/bwmarrin/snowflake",
        sum = "h1:lTJlWdyhwqq7h29GtuIDHW/xi+sMN+JOLMgYAwQ5O74=",
        version = "v0.0.0-20180412010544-68117e6bbede",
    )

    go_repository(
        name = "com_github_davecgh_go_spew",
        importpath = "github.com/davecgh/go-spew",
        sum = "h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=",
        version = "v1.1.1",
    )

    go_repository(
        name = "com_github_deckarep_golang_set",
        importpath = "github.com/deckarep/golang-set",
        sum = "h1:njG8LmGD6JCWJu4bwIKmkOHvch70UOEIqczl5vp7Gok=",
        version = "v0.0.0-20171013212420-1d4478f51bed",
    )

    go_repository(
        name = "com_github_fsnotify_fsnotify",
        importpath = "github.com/fsnotify/fsnotify",
        sum = "h1:IXs+QLmnXW2CcXuY+8Mzv/fWEsPGWxqefPtCP5CnV9I=",
        version = "v1.4.7",
    )

    go_repository(
        name = "com_github_ghodss_yaml",
        importpath = "github.com/ghodss/yaml",
        sum = "h1:wQHKEahhL6wmXdzwWG11gIVCkOv05bNOh+Rxn0yngAk=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_go_yaml_yaml",
        importpath = "github.com/go-yaml/yaml",
        sum = "h1:RYi2hDdss1u4YE7GwixGzWwVo47T8UQwnTLB6vQiq+o=",
        version = "v2.1.0+incompatible",
    )

    go_repository(
        name = "com_github_gogo_protobuf",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gogo/protobuf",
        sum = "h1:2jyBKDKU/8v3v2xVR2PtiWQviFUyiaGk2rpfyFT8rTM=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_golang_glog",
        importpath = "github.com/golang/glog",
        sum = "h1:VKtxabqXZkF25pY9ekfRL6a582T4P37/31XEstQ5p58=",
        version = "v0.0.0-20160126235308-23def4e6c14b",
    )

    go_repository(
        name = "com_github_golang_lint",
        importpath = "github.com/golang/lint",
        sum = "h1:r7LK7GmaeiuSi9Jy7FbeMNoB83lsf7uo7xgmWNrg9YE=",
        version = "v0.0.0-20180319214916-85993ffd0a6c",
    )

    go_repository(
        name = "com_github_golang_protobuf",
        importpath = "github.com/golang/protobuf",
        sum = "h1:P3YflyNX/ehuJFLhxviNdFxQPkGK5cDcApsge1SqnvM=",
        version = "v1.2.0",
    )

    go_repository(
        name = "com_github_google_btree",
        importpath = "github.com/google/btree",
        sum = "h1:964Od4U6p2jUkFxvCydnIczKteheJEzHRToSGK3Bnlw=",
        version = "v0.0.0-20180813153112-4030bb1f1f0c",
    )

    go_repository(
        name = "com_github_google_go_cmp",
        importpath = "github.com/google/go-cmp",
        sum = "h1:+dTQ8DZQJz0Mb/HjFlkptS1FeQ4cWSnN941F8aEG4SQ=",
        version = "v0.2.0",
    )

    go_repository(
        name = "com_github_google_go_github",
        importpath = "github.com/google/go-github",
        sum = "h1:a35eq0ruIMahsuI5Vc/S554/B+JTG+ZOxz0h+b7v4Qw=",
        version = "v0.0.0-20171108000855-8c08f4fba5e0",
    )

    go_repository(
        name = "com_github_google_go_querystring",
        importpath = "github.com/google/go-querystring",
        sum = "h1:zLTLjkaOFEFIOxY5BWLFLwh+cL8vOBW4XJ2aqLE/Tf0=",
        version = "v0.0.0-20170111101155-53e6ce116135",
    )

    go_repository(
        name = "com_github_google_gofuzz",
        importpath = "github.com/google/gofuzz",
        sum = "h1:+RRA9JqSOZFfKrOeqr2z77+8R2RKyh8PG66dcu1V0ck=",
        version = "v0.0.0-20170612174753-24818f796faf",
    )

    go_repository(
        name = "com_github_googleapis_gax_go",
        importpath = "github.com/googleapis/gax-go",
        sum = "h1:j0GKcs05QVmm7yesiZq2+9cxHkNK9YM6zKx4D2qucQU=",
        version = "v2.0.0+incompatible",
    )

    go_repository(
        name = "com_github_googleapis_gnostic",
        importpath = "github.com/googleapis/gnostic",
        sum = "h1:l6N3VoaVzTncYYW+9yOz2LJJammFZGBO13sqgEhpy9g=",
        version = "v0.2.0",
    )

    go_repository(
        name = "com_github_gorilla_context",
        importpath = "github.com/gorilla/context",
        sum = "h1:9oNbS1z4rVpbnkHBdPZU4jo9bSmrLpII768arSyMFgk=",
        version = "v0.0.0-20160226214623-1ea25387ff6f",
    )

    go_repository(
        name = "com_github_gorilla_securecookie",
        importpath = "github.com/gorilla/securecookie",
        sum = "h1:miw7JPhV+b/lAHSXz4qd/nN9jRiAFV5FwjeKyCS8BvQ=",
        version = "v1.1.1",
    )

    go_repository(
        name = "com_github_gorilla_sessions",
        importpath = "github.com/gorilla/sessions",
        sum = "h1:OuuPl66BpF1q3OEkaPpp+VfzxrBBY62ATGdWqql/XX8=",
        version = "v0.0.0-20160922145804-ca9ada445741",
    )

    go_repository(
        name = "com_github_gregjones_httpcache",
        importpath = "github.com/gregjones/httpcache",
        sum = "h1:ShTPMJQes6tubcjzGMODIVG5hlrCeImaBnZzKF2N8SM=",
        version = "v0.0.0-20181110185634-c63ab54fda8f",
    )

    go_repository(
        name = "com_github_hashicorp_errwrap",
        importpath = "github.com/hashicorp/errwrap",
        sum = "h1:prjrVgOk2Yg6w+PflHoszQNLTUh4kaByUcEWM/9uin4=",
        version = "v0.0.0-20141028054710-7554cd9344ce",
    )

    go_repository(
        name = "com_github_hashicorp_go_multierror",
        importpath = "github.com/hashicorp/go-multierror",
        sum = "h1:RUacJnONqfKgDeok3I3IqMa8e5+B3qzBIbNK4dZK65k=",
        version = "v0.0.0-20170622060955-83588e72410a",
    )

    go_repository(
        name = "com_github_hpcloud_tail",
        importpath = "github.com/hpcloud/tail",
        sum = "h1:nfCOvKYfkgYP8hkirhJocXT2+zOD8yUNjXaWfTlyFKI=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_imdario_mergo",
        importpath = "github.com/imdario/mergo",
        sum = "h1:xTNEAn+kxVO7dTZGu0CegyqKZmoWFI0rF8UxjlB2d28=",
        version = "v0.3.6",
    )

    go_repository(
        name = "com_github_json_iterator_go",
        importpath = "github.com/json-iterator/go",
        sum = "h1:gL2yXlmiIo4+t+y32d4WGwOjKGYcGOuyrg46vadswDE=",
        version = "v1.1.5",
    )

    go_repository(
        name = "com_github_knative_build",
        importpath = "github.com/knative/build",
        sum = "h1:Zu+P0p8aq+EIeA8fyTzFX7kX7xY3ip+GrYPPkP/HhTI=",
        version = "v0.1.0",
    )

    go_repository(
        name = "com_github_kr_pretty",
        importpath = "github.com/kr/pretty",
        sum = "h1:L/CwN0zerZDmRFUapSPitk6f+Q3+0za1rQkzVuMiMFI=",
        version = "v0.1.0",
    )

    go_repository(
        name = "com_github_kr_pty",
        importpath = "github.com/kr/pty",
        sum = "h1:VkoXIwSboBpnk99O/KFauAEILuNHv5DVFKZMBN/gUgw=",
        version = "v1.1.1",
    )

    go_repository(
        name = "com_github_kr_text",
        importpath = "github.com/kr/text",
        sum = "h1:45sCR5RtlFHMR4UwH9sdQ5TC8v0qDQCHnXt+kaKSTVE=",
        version = "v0.1.0",
    )

    go_repository(
        name = "com_github_mattn_go_zglob",
        importpath = "github.com/mattn/go-zglob",
        sum = "h1:+/5w9rBfwLFufeFqRfnHfBiWJrBoCKDgAESq6qELHmM=",
        version = "v0.0.0-20180627001149-c436403c742d",
    )

    go_repository(
        name = "com_github_matttproud_golang_protobuf_extensions",
        importpath = "github.com/matttproud/golang_protobuf_extensions",
        sum = "h1:YNOwxxSJzSUARoD9KRZLzM9Y858MNGCOACTvCW9TSAc=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_modern_go_concurrent",
        importpath = "github.com/modern-go/concurrent",
        sum = "h1:TRLaZ9cD/w8PVh93nsPXa1VrQ6jlwL5oN8l14QlcNfg=",
        version = "v0.0.0-20180306012644-bacd9c7ef1dd",
    )

    go_repository(
        name = "com_github_modern_go_reflect2",
        importpath = "github.com/modern-go/reflect2",
        sum = "h1:9f412s+6RmYXLWZSEzVVgPGK7C2PphHj5RJrvfx9AWI=",
        version = "v1.0.1",
    )

    go_repository(
        name = "com_github_onsi_ginkgo",
        importpath = "github.com/onsi/ginkgo",
        sum = "h1:WSHQ+IS43OoUrWtD1/bbclrwK8TTH5hzp+umCiuxHgs=",
        version = "v1.7.0",
    )

    go_repository(
        name = "com_github_onsi_gomega",
        importpath = "github.com/onsi/gomega",
        sum = "h1:RE1xgDvH7imwFD45h+u2SgIfERHlS2yNG4DObb5BSKU=",
        version = "v1.4.3",
    )

    go_repository(
        name = "com_github_peterbourgon_diskv",
        importpath = "github.com/peterbourgon/diskv",
        sum = "h1:UBdAOUP5p4RWqPBg048CAvpKN+vxiaj6gdUUzhl4XmI=",
        version = "v2.0.1+incompatible",
    )

    go_repository(
        name = "com_github_pmezard_go_difflib",
        importpath = "github.com/pmezard/go-difflib",
        sum = "h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_prometheus_client_golang",
        importpath = "github.com/prometheus/client_golang",
        sum = "h1:1921Yw9Gc3iSc4VQh3PIoOqgPCZS7G/4xQNVUp8Mda8=",
        version = "v0.8.0",
    )

    go_repository(
        name = "com_github_prometheus_client_model",
        importpath = "github.com/prometheus/client_model",
        sum = "h1:13pIdM2tpaDi4OVe24fgoIS7ZTqMt0QI+bwQsX5hq+g=",
        version = "v0.0.0-20170216185247-6f3806018612",
    )

    go_repository(
        name = "com_github_prometheus_common",
        importpath = "github.com/prometheus/common",
        sum = "h1:g2v6dZgmqj2wYGPgHYX5WVaQ9IwV1ylsSiD+f8RvS1Y=",
        version = "v0.0.0-20171104095907-e3fb1a1acd76",
    )

    go_repository(
        name = "com_github_prometheus_procfs",
        importpath = "github.com/prometheus/procfs",
        sum = "h1:leRfx9kcgnSDkqAFhaaUcRqpAZgnFdwZkZcdRcea1h0=",
        version = "v0.0.0-20171017214025-a6e9df898b13",
    )

    go_repository(
        name = "com_github_satori_go_uuid",
        importpath = "github.com/satori/go.uuid",
        sum = "h1:0uYX9dsZ2yD7q2RtLRtPSdGDWzjeM3TbMJP9utgA0ww=",
        version = "v1.2.0",
    )

    go_repository(
        name = "com_github_shurcool_githubv4",
        importpath = "github.com/shurcooL/githubv4",
        sum = "h1:cppRIvEpuZcSdhbhyJZ/3ThCPYlx6xuZg8Qid/0+bz0=",
        version = "v0.0.0-20180925043049-51d7b505e2e9",
    )

    go_repository(
        name = "com_github_shurcool_go",
        importpath = "github.com/shurcooL/go",
        sum = "h1:WqsobZNyIWv3xVI0pgkziPvpqny4wZvufTBYPzOlQNw=",
        version = "v0.0.0-20180410215514-47fa5b7ceee6",
    )

    go_repository(
        name = "com_github_shurcool_graphql",
        importpath = "github.com/shurcooL/graphql",
        sum = "h1:g7taRn8UFRnLp1YqHP+ScI5hc1+2P/jrBm8wpsG9Ies=",
        version = "v0.0.0-20180302221403-3d276b9dcc6b",
    )

    go_repository(
        name = "com_github_sirupsen_logrus",
        importpath = "github.com/sirupsen/logrus",
        sum = "h1:gzbtLsZC3Ic5PptoRG+kQj4L60qjK7H7XszrU163JNQ=",
        version = "v1.0.4",
    )

    go_repository(
        name = "com_github_spf13_pflag",
        importpath = "github.com/spf13/pflag",
        sum = "h1:aCvUg6QPl3ibpQUxyLkrEkCHtPqYJL4x9AuhqVqFis4=",
        version = "v1.0.1",
    )

    go_repository(
        name = "com_github_stretchr_objx",
        importpath = "github.com/stretchr/objx",
        sum = "h1:4G4v2dO3VZwixGIRoQ5Lfboy6nUhCyYzaqnIAPPhYs4=",
        version = "v0.1.0",
    )

    go_repository(
        name = "com_github_stretchr_testify",
        importpath = "github.com/stretchr/testify",
        sum = "h1:TivCn/peBQ7UY8ooIcPgZFpTNSz0Q2U6UrFlUfqbe0Q=",
        version = "v1.3.0",
    )

    go_repository(
        name = "com_google_cloud_go",
        importpath = "cloud.google.com/go",
        sum = "h1:w1svupRqvZnfjN9+KksMiggoIRQuMzWkVzpxcR96xDs=",
        version = "v0.23.0",
    )

    go_repository(
        name = "in_gopkg_airbrake_gobrake_v2",
        importpath = "gopkg.in/airbrake/gobrake.v2",
        sum = "h1:7z2uVWwn7oVeeugY1DtlPAy5H+KYgB1KeKTnqjNatLo=",
        version = "v2.0.9",
    )

    go_repository(
        name = "in_gopkg_check_v1",
        importpath = "gopkg.in/check.v1",
        sum = "h1:qIbj1fsPNlZgppZ+VLlY7N33q108Sa+fhmuc+sWQYwY=",
        version = "v1.0.0-20180628173108-788fd7840127",
    )

    go_repository(
        name = "in_gopkg_fsnotify_v1",
        importpath = "gopkg.in/fsnotify.v1",
        sum = "h1:xOHLXZwVvI9hhs+cLKq5+I5onOuwQLhQwiu63xxlHs4=",
        version = "v1.4.7",
    )

    go_repository(
        name = "in_gopkg_gemnasium_logrus_airbrake_hook_v2",
        importpath = "gopkg.in/gemnasium/logrus-airbrake-hook.v2",
        sum = "h1:OAj3g0cR6Dx/R07QgQe8wkA9RNjB2u4i700xBkIT4e0=",
        version = "v2.1.2",
    )

    go_repository(
        name = "in_gopkg_inf_v0",
        importpath = "gopkg.in/inf.v0",
        sum = "h1:73M5CoZyi3ZLMOyDlQh031Cx6N9NDJ2Vvfl76EDAgDc=",
        version = "v0.9.1",
    )

    go_repository(
        name = "in_gopkg_robfig_cron_v2",
        importpath = "gopkg.in/robfig/cron.v2",
        sum = "h1:E846t8CnR+lv5nE+VuiKTDG/v1U2stad0QzddfJC7kY=",
        version = "v2.0.0-20150107220207-be2e0b0deed5",
    )

    go_repository(
        name = "in_gopkg_tomb_v1",
        importpath = "gopkg.in/tomb.v1",
        sum = "h1:uRGJdciOHaEIrze2W8Q3AKkepLTh2hOroT7a+7czfdQ=",
        version = "v1.0.0-20141024135613-dd632973f1e7",
    )

    go_repository(
        name = "in_gopkg_yaml_v2",
        importpath = "gopkg.in/yaml.v2",
        sum = "h1:mUhvW9EsL+naU5Q3cakzfE91YhliOondGd6ZrsDBHQE=",
        version = "v2.2.1",
    )

    go_repository(
        name = "io_k8s_api",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/api",
        sum = "h1:iGq7zEPXFb0IeXAQK5RiYT1SVKX/af9F9Wv0M+yudPY=",
        version = "v0.0.0-20181221193117-173ce66c1e39",
    )

    go_repository(
        name = "io_k8s_apiextensions_apiserver",
        importpath = "k8s.io/apiextensions-apiserver",
        sum = "h1:qzRasX5+11GdmU3I4NOunPRJva2BB6YsE3lQ7RWXXN0=",
        version = "v0.0.0-20190119024419-80a4532647cb",
    )

    go_repository(
        name = "io_k8s_apimachinery",
        build_file_generation = "on",
        build_file_name = "BUILD.bazel",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/apimachinery",
        sum = "h1:CsgbEA8905OlpVLNKWD4GacPex50kFbqhotVNPew+dU=",
        version = "v0.0.0-20180228050457-302974c03f7e",
    )

    go_repository(
        name = "io_k8s_client_go",
        importpath = "k8s.io/client-go",
        sum = "h1:F1IqCqw7oMBzDkqlcBymRq1450wD0eNqLE9jzUrIi34=",
        version = "v10.0.0+incompatible",
    )

    go_repository(
        name = "io_k8s_klog",
        importpath = "k8s.io/klog",
        sum = "h1:I5HMfc/DtuVaGR1KPwUrTc476K8NCqNBldC7H4dYEzk=",
        version = "v0.1.0",
    )

    go_repository(
        name = "io_k8s_sigs_yaml",
        importpath = "sigs.k8s.io/yaml",
        sum = "h1:4A07+ZFc2wgJwo8YNlQpr1rVlgUDlxXHhPJciaPY5gs=",
        version = "v1.1.0",
    )

    go_repository(
        name = "io_k8s_test_infra",
        importpath = "k8s.io/test-infra",
        sum = "h1:rUdojAZPv4qa47cBk0DJuNxyEeL6dMSyh1t3RrT/swI=",
        version = "v0.0.0-20181127230316-bfbc61258394",
    )

    go_repository(
        name = "io_opencensus_go",
        importpath = "go.opencensus.io",
        sum = "h1:QKHMRkzPeS4VokdZHF+viaX1F5vqbsAEiH7zEjuf71M=",
        version = "v0.13.0",
    )

    go_repository(
        name = "org_golang_google_api",
        importpath = "google.golang.org/api",
        sum = "h1:P8qjZSvfkGEHH+N+DELB1pFvDdA7SJ7x5pjezw/mZH0=",
        version = "v0.0.0-20180131010904-ffa5046912fd",
    )

    go_repository(
        name = "org_golang_google_appengine",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/appengine",
        sum = "h1:dN4LljjBKVChsv0XCSI+zbyzdqrkEwX5LQFUMRSGqOc=",
        version = "v1.0.0",
    )

    go_repository(
        name = "org_golang_google_genproto",
        importpath = "google.golang.org/genproto",
        sum = "h1:sYGAnR6gvwiXWfMXJiLUXtU2C3/03O+RRE8VbWJmM0E=",
        version = "v0.0.0-20171103030625-11c7f9e547da",
    )

    go_repository(
        name = "org_golang_google_grpc",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/grpc",
        sum = "h1:Vw1JtR07h6jezLtFKVRNMq5BGqECN1y9dPSEM5f+f7s=",
        version = "v1.7.2",
    )

    go_repository(
        name = "org_golang_x_crypto",
        importpath = "golang.org/x/crypto",
        sum = "h1:OfaUle5HH9Y0obNU74mlOZ/Igdtwi3eGOKcljJsTnbw=",
        version = "v0.0.0-20180214000028-650f4a345ab4",
    )

    go_repository(
        name = "org_golang_x_net",
        importpath = "golang.org/x/net",
        sum = "h1:nTDtHvHSdCn1m6ITfMRqtOd/9+7a3s8RBNOZ3eYZzJA=",
        version = "v0.0.0-20180906233101-161cd47e91fd",
    )

    go_repository(
        name = "org_golang_x_oauth2",
        importpath = "golang.org/x/oauth2",
        sum = "h1:nP0LlV1P7+z/qtbjHygz+Bba7QsbB4MqdhGJmAyicuI=",
        version = "v0.0.0-20171106152852-9ff8ebcc8e24",
    )

    go_repository(
        name = "org_golang_x_sync",
        importpath = "golang.org/x/sync",
        sum = "h1:wMNYb4v58l5UBM7MYRLPG6ZhfOqbKu7X5eyFl8ZhKvA=",
        version = "v0.0.0-20180314180146-1d60e4601c6f",
    )

    go_repository(
        name = "org_golang_x_sys",
        importpath = "golang.org/x/sys",
        sum = "h1:o3PsSEY8E4eXWkXrIP9YJALUkVZqzHJT5DOasTyn8Vs=",
        version = "v0.0.0-20180909124046-d0be0721c37e",
    )

    go_repository(
        name = "org_golang_x_text",
        importpath = "golang.org/x/text",
        sum = "h1:g61tztE5qeGQ89tm6NTjjM9VPIm088od1l6aSorWRWg=",
        version = "v0.3.0",
    )

    go_repository(
        name = "org_golang_x_time",
        importpath = "golang.org/x/time",
        sum = "h1:fqgJT0MGcGpPgpWU7VRdRjuArfcOvC4AoJmILihzhDg=",
        version = "v0.0.0-20181108054448-85acf8d2951c",
    )

    go_repository(
        name = "org_golang_x_tools",
        importpath = "golang.org/x/tools",
        sum = "h1:aGoTCTvIjOOSnuAeiybrXAIbqAuPVB4DvRXlCrKT02s=",
        version = "v0.0.0-20190121143147-24cd39ecf745",
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
        commit = "8e66885c52b0",
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
        tag = "v1.2.0",
    )

    go_repository(
        name = "com_github_kisielk_gotool",
        commit = "0de1eaf82fa3",
        importpath = "github.com/kisielk/gotool",
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
        commit = "8f45f776aaf1",
        importpath = "golang.org/x/lint",
    )
    go_repository(
        name = "com_github_bazelbuild_buildtools",
        importpath = "github.com/bazelbuild/buildtools",
        sum = "h1:VuTBHPJNCQ88Okm9ld5SyLCvU50soWJYQYjQFdcDxew=",
        version = "v0.0.0-20180226164855-80c7f0d45d7e",
    )

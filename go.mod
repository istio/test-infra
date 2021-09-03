module istio.io/test-infra

go 1.12

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.2.0+incompatible
	github.com/garyburd/redigo => github.com/garyburd/redigo v1.6.0 // for LICENSE
	github.com/otiai10/curr => github.com/otiai10/curr v1.0.0 // for LICENSE
	github.com/otiai10/mint => github.com/otiai10/mint v1.3.0 // remove dependency on bou.ke/monkey
	github.com/pelletier/go-buffruneio => github.com/pelletier/go-buffruneio v0.2.1-0.20190103235659-25c428535bd3 // for LICENSE
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.7.0-alpha.6.0.20201106193838-8d0107636985
)

require (
	cloud.google.com/go/storage v1.12.0
	github.com/ghodss/yaml v1.0.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-cmp v0.5.5
	github.com/google/go-github v17.0.0+incompatible
	github.com/hashicorp/go-multierror v1.1.0
	github.com/kr/pretty v0.2.0
	github.com/prometheus/client_golang v1.11.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/ulule/deepcopier v0.0.0-20200430083143-45decc6639b6
	golang.org/x/net v0.0.0-20210428140749-89ef3d95e781
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	google.golang.org/api v0.32.0
	gopkg.in/robfig/cron.v2 v2.0.0-20150107220207-be2e0b0deed5
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/test-infra v0.0.0-20210719201052-da1e7f5b5334
	sigs.k8s.io/boskos v0.0.0-20210521134047-36bb085667e7
	sigs.k8s.io/yaml v1.2.0
)

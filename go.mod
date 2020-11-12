module istio.io/test-infra

go 1.12

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	github.com/garyburd/redigo => github.com/garyburd/redigo v1.6.0 // for LICENSE
	github.com/otiai10/curr => github.com/otiai10/curr v1.0.0 // for LICENSE
	github.com/otiai10/mint => github.com/otiai10/mint v1.3.0 // remove dependency on bou.ke/monkey
	github.com/pelletier/go-buffruneio => github.com/pelletier/go-buffruneio v0.2.1-0.20190103235659-25c428535bd3 // for LICENSE
	k8s.io/api => k8s.io/api v0.17.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/client-go => k8s.io/client-go v0.17.3
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.5.0
)

require (
	cloud.google.com/go/storage v1.1.2
	github.com/ghodss/yaml v1.0.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-cmp v0.4.0
	github.com/google/go-github v17.0.0+incompatible
	github.com/hashicorp/go-multierror v1.0.0
	github.com/kr/pretty v0.2.0
	github.com/prometheus/client_golang v1.5.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/api v0.15.0
	gopkg.in/robfig/cron.v2 v2.0.0-20150107220207-be2e0b0deed5
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/test-infra v0.0.0-20200514184223-ba32c8aae783
	sigs.k8s.io/boskos v0.0.0-20200530174753-71e795271860
	sigs.k8s.io/yaml v1.2.0
)

module istio.io/test-infra

go 1.12

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	github.com/otiai10/curr => github.com/otiai10/curr v1.0.0 // for LICENSE
	github.com/otiai10/mint => github.com/otiai10/mint v1.3.0 // remove dependency on bou.ke/monkey
	github.com/pelletier/go-buffruneio => github.com/pelletier/go-buffruneio v0.2.1-0.20190103235659-25c428535bd3 // for LICENSE
	k8s.io/api => k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
)

require (
	cloud.google.com/go/storage v1.1.2
	github.com/ghodss/yaml v1.0.0
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-github v17.0.0+incompatible
	github.com/hashicorp/go-multierror v1.0.0
	github.com/kr/pretty v0.1.0
	github.com/prometheus/client_golang v1.5.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/api v0.11.0
	gopkg.in/robfig/cron.v2 v2.0.0-20150107220207-be2e0b0deed5
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/test-infra v0.0.0-20200304232039-27a0743d7806
	sigs.k8s.io/yaml v1.1.0
)

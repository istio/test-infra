module istio.io/test-infra

go 1.16

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.2.0+incompatible
	k8s.io/api => k8s.io/api v0.22.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.2
	k8s.io/client-go => k8s.io/client-go v0.22.2
)

require (
	cloud.google.com/go/storage v1.18.2
	github.com/ghodss/yaml v1.0.0
	github.com/golang/glog v1.0.0
	github.com/google/go-cmp v0.5.6
	github.com/google/go-github v17.0.0+incompatible
	github.com/hashicorp/go-multierror v1.1.1
	github.com/prometheus/client_golang v1.11.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20220107192237-5cfca573fb4d
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/api v0.64.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/test-infra v0.0.0-20220110151312-600d25dbe068
	sigs.k8s.io/boskos v0.0.0-20211118173702-344faec9d22a
	sigs.k8s.io/yaml v1.3.0
)

require (
	cloud.google.com/go/container v1.0.0 // indirect
	cloud.google.com/go/monitoring v1.2.0 // indirect
	cloud.google.com/go/trace v1.0.0 // indirect
)

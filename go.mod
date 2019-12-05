module istio.io/test-infra

go 1.12

replace (
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
	github.com/prometheus/client_golang v1.2.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20191021144547-ec77196f6094
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/api v0.11.0
	gopkg.in/yaml.v2 v2.2.7
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/test-infra v0.0.0-20191205175432-7e847bca3c74
	sigs.k8s.io/yaml v1.1.0
)

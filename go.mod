module istio.io/test-infra

go 1.12

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
)

require (
	cloud.google.com/go v0.44.3
	github.com/ghodss/yaml v1.0.0
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-github v17.0.0+incompatible
	github.com/hashicorp/go-multierror v0.0.0-20171204182908-b7773ae21874
	github.com/kr/pretty v0.1.0
	github.com/prometheus/client_golang v0.9.4
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.5
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/tools v0.0.0-20190731214159-1e85ed8060aa // indirect
	google.golang.org/api v0.9.0
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/test-infra v0.0.0-20191014212758-9a9c89077c13
	sigs.k8s.io/yaml v1.1.0
)

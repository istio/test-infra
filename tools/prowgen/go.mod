module istio.io/test-infra/tools/prowgen

go 1.16

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.2.0+incompatible
	k8s.io/api => k8s.io/api v0.22.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.2
	k8s.io/client-go => k8s.io/client-go v0.22.2
)

require (
	github.com/google/go-cmp v0.5.6
	github.com/hashicorp/go-multierror v1.1.1
	github.com/imdario/mergo v0.3.12
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	gopkg.in/robfig/cron.v2 v2.0.0-20150107220207-be2e0b0deed5
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/test-infra v0.0.0-20220110151312-600d25dbe068
	sigs.k8s.io/yaml v1.3.0
)

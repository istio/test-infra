# istio-builder

Istio builder is used to build release artifacts from Cloud Builder.

The istio-builder can be added to your ```${PATH}``` to get you started. It includes all the tools you need to build istio

```

$ wget https://raw.githubusercontent.com/istio/test-infra/master/docker/istio_builders/istio_builder/istio-builder\
  -O ${HOME}/bin/istio-builder
$ chmod a+x ${HOME}/bin/istio-builder

```

From your istio checkout

```
$ istio-builder make build
```

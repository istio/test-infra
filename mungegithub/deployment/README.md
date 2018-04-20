## Update Submit Queue
* Build mungegithub binary
  
  Checkout kubernetes/test-infra repo. 
  
  >Don't build directly from kubernetes/test-infra master unless you understand what major new features
  are involved. Currently version is built from github.com/yutongz/k8s-test-infra/tree/master
  
  On repo root directory
  ```bash
  $ bazel build //mungegithub:mungegithub
  $ cp bazel-bin/mungegithub/mungegithub mungegithub/
  ```

* Build and push docker image

  Make sure you are authorized to push to that docker hub, here is the hub and tag we are using:
  ```bash
  $ cd mungegithub
  $ CONTAINER=gcr.io/istio-testing/mungegithub/submit-queue:Sep-11-yutongz-a81300506273c9c27bc6fcd33a4b12cf0feace69
  $ docker build --pull -t $(CONTAINER) -f Dockerfile-submit-queue .
  $ docker push $(CONTAINER)
  ```
  
* Re-deploy Submit Queue instence

  For example, if you want to deploy for proxy:

  ```bash
  $ REPO=proxy make deploy
  ```
  
## Change config file

  Pull the latest istio/test-infra, change the configmap.yaml under your target repo directory and go back to `test-infra/mungegithub`
 
 ```bash
 $ REPO=proxy make update-config 
 ```
  It will restart the Submit Queue for that repo to make sure it syncs with the new config.
 

## Host Submit Queue

We are using one basic load-balancer for each Submit Queue.

* Auth: http://35.197.10.29:8080
* Istio: http://35.197.104.17:8080
* Mixer: http://35.197.95.47:8080
* Pilot: http://35.199.174.118:8080
* Test-Infra: http://104.196.250.254:8080
* Mixerclient: http://35.199.147.28:8080
* Proxy: http://35.197.99.161:8080


## Maintain Submit Queue

Most maintainance only involves changes in configmap. Since k8s currently doesn't support `kubectl replace configmap`,
there is [a open issue](https://github.com/kubernetes/kubernetes/issues/30558). When you need to update configmap,
you need to do:

```bash
$ kubectl delete configmap istio-sq-config
$ kubectl create configmap istio-sq-config --from-file=mungegithub/deployment/istio/configmap.yaml
```

* Required CI tests change
  
  Change `required-contexts` and `required-retest-contexts`. Usually they should be the same.
  
* Emergency stop submit-queue

```bash
$ kubectl delete deployment istio-submit-queue
```

* Enable/disable "approve" gating

  Change `gate-approved`.

* Frequence change

  Change `period`
  

## Troubleshooting

> Make sure read [user manual](https://github.com/istio/test-infra/blob/master/mungegithub/README.md) first.

* ~~Someone comment "/lgtm", but istio-testing didn't add "lgtm" label~~

  ~~Github webhook issue. Simply comment "/lgtm" again.~~
  [FIXED]
  
* ~~CI already finished successfully, but Submit Queue still complains one is not green~~
  
  ~~Believe it's the webhook issue too. CI status change will not trigger Submit Queue, you need to 
  punch it by leaving any comment. [Example](https://github.com/istio/istio/pull/730)~~
  
  [FIXED]
  
* Check log

  ```bash
  $ kubectl logs istio-submit-queue-2564258885-qb09l
  ```
  

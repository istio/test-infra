Enable RBE on your workstation:

```sh
curl https://raw.githubusercontent.com/kubernetes/test-infra/master/rbe/configure.sh -o configure.sh
chmod configure.sh
./configure.sh istio-testing # or some other project
bazel test --config=remote-istio-testing //...
```

How can I always use RBE?

```sh
# vim ~/.bazelrc
build --config=remote-istio-testing
```

Create/update a RBE cluster:

```sh
# cp install-istio-testing.sh install-into-my-project.sh
# edit values in this file
./install-istio-testing.sh
```

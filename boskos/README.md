# Boskos
-----

## Background

Boskos is a resource manager service, that handles and manages different kind of
resources and transition between different states.

In istio, we are using Boskos to pre-provision cluster and vms on GCP. This
functionality was added with the Mason module.

## Setup

Boskos runs on a kubernetes cluster. Initially only Prow jobs will be using
Boskos. Since Boskos service is only available intra cluster, it is important
that Boskos runs in the same cluster where test jobs are running.

### Deployments

1. [boskos](cluster/boskos-deployment.yaml)     - Main component, handles and manage resource
2. [janitor](cluster/janitor-deployment.yaml)   - Clean dirty GCP project to free state
3. [reaper](cluster/reaper-deployment.yaml)     - Look for resources that
   are owned and not being updated, and mark them as dirty
4. [mason](cluster/mason-deployment.yaml)       - Transform dirty mason resources to
   free

Boskos is using boskos@istio-testing.iam.gserviceaccount.com from the
istio-testing project. For first time setup, one needs to create a new
key from the UI and download it locally, and then run

```bash
make init boskos-config deploy SERVICE-ACCOUNT-JSON={path to service account json file}
```

### Upgrade

boskos, janitor, reaper images are coming from
[kubernetes/test-infra](https://github.com/kubernetes/test-infra/tree/master/boskos),
so we should only use those images. However for mason, we need to add our own
resources implementations.

In order to create to update mason deployment run the following:

```bash
make mason-image

# Update `boskos/cluster/mason-deployment.yaml` to point to the new image, create a PR, then:
make deploy
```

## Mason

Mason is used for resources that have resources dependency. Mason resource are
virtual, they only exist inside other resources. An example of a mason resource
can be a cluster, which needs to exist in a GCP project. In order to create a
cluster you need a GCP project first.

Mason is configured in the `resources.yaml` file. A typical mason config would look like this:

```yaml
resources:
...
- name: gke-e2e-test
  state: dirty
  min-count: 10
  max-count: 50
  needs:
    gcp-project: 1
  config:
    type: GCPResourceConfig
    content: |
      gcp-project:
        - clusters:
          - machinetype: n1-standard-4
            numnodes: 5
            version: 1.10
            scopes:
            - https://www.googleapis.com/auth/cloud-platform
            - https://www.googleapis.com/auth/trace.append
          - machinetype: n1-standard-4
            numnodes: 5
            version: 1.10
            scopes:
            - https://www.googleapis.com/auth/cloud-platform
            - https://www.googleapis.com/auth/trace.append
          vms:
          - machinetype: n1-standard-4
            sourceimage: projects/debian-cloud/global/images/debian-9-stretch-v20180206
            tags:
            - http-server
            - https-server
            scopes:
            - https://www.googleapis.com/auth/cloud-platform
            - https://www.googleapis.com/auth/trace.append
```

The name points to the resource types from the `gcp-project` in `resources.yaml`

```yaml
resources:
...
- type: gcp-project
  state: dirty
  names:
  - sebvas-boskos-01
  - sebvas-boskos-02
  - sebvas-boskos-03
  - istio-boskos-01
  - istio-boskos-02
  - istio-boskos-03
  - istio-boskos-04
  - istio-boskos-05
  - istio-boskos-06
  - istio-boskos-07
```

As you can see, a `gke-e2e-test` resource is composed of 1 `gcp-project` resource
(needs). And we'll use the `GCPResourceConfig` type to parse the config (content)
and create the resource (in that case a VM and a cluster)


## Adding new resources

The number of real resources should be greater than equal to the virtual
resource needs. Any update that does not satisfy this constraint will fail the
test.

### Adding a new GCP project

When adding new GCP project, we need to make sure that the following user are
owner of the project:

1. boskos@istio-testing.iam.gserviceaccount.com
2. istio-prow-test-job@istio-testing.iam.gserviceaccount.com
3. mdb.istio-testing@google.com

and that container and compute api are enabled. It is recommended to create the
project from the UI and to link a billing account from there. Once that's done
the following script can help

```bash
./set-boskos-project.sh -p {project id}
```

We can then add a new line in the gcp section of resources.yaml

### Adding a new virtual resource in an existing pool

For each new virtual resource, we need to add the associated physical resouces.
Once that's done, the resources.yaml file can be updated with a new line.

### Updating configurations

In order to update resources and mason configs, you can run

```bash
make boskos-config
```

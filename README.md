## Table of Content

- [Deployment information](#deployment-information)
- [Kubernetes Setup](#kubernetes-setup)
  * [Creating Jenkins Cluster](#creating-jenkins-cluster)
  * [Setting up Jenkins](#setting-up-jenkins)
    + [Setup AppEngine Reverse Proxy](#setup-appengine-reverse-proxy)
    + [Jenkins Environment variables](#jenkins-environment-variables)
  * [Updating Jenkins to point to the cluster](#updating-jenkins-to-point-to-the-cluster)
  * [Creating an Hazelcast cluster](#creating-an-hazelcast-cluster)
- [Maintenance](#maintenance)
  * [Keeping Jenkins Plugins up to date](#keeping-jenkins-plugins-up-to-date)
- [Troubleshooting](#troubleshooting)
  * [All Slaves are marked as offline](#all-slaves-are-marked-as-offline)

## Deployment information ##

    $ export PROJECT_ID='istio-testing'
    $ export ZONE='us-central1-f'
    $ export CLUSTER_NAME='jenkins-cluster'
    $ export K8S_SCOPES='https://www.googleapis.com/auth/appengine.admin,
    https://www.googleapis.com/auth/cloud-platform,
    https://www.googleapis.com/auth/compute,
    https://www.googleapis.com/auth/devstorage.full_control,
    https://www.googleapis.com/auth/devstorage.read_only,
    https://www.googleapis.com/auth/logging.write,
    https://www.googleapis.com/auth/projecthosting,
    https://www.googleapis.com/auth/servicecontrol,
    https://www.googleapis.com/auth/service.management'

## Kubernetes Setup ##

### Creating Jenkins Cluster ###

Jenkins runs all its slaves in a Kubernetes Cluster. The cluster name is
hardcoded in the Jenkinsfile.

    $ gcloud container \
    --project "${PROJECT_ID}" \
    clusters create "${CLUSTER_NAME}" \
    --zone "${ZONE}" \
    --machine-type "n1-highmem-32" \
    --disk-size 400 \
    --scopes "${K8S_SCOPES}" \
    --num-nodes "2" \
    --image-type=CONTAINER_VM \
    --network "default" \
    --enable-cloud-logging \
    --no-enable-cloud-monitoring

    # Update kubectl config.
    # You might want to run this command as well on your desktop.
    $ gcloud container clusters get-credentials "${CLUSTER_NAME}" \
    --project "${PROJECT_ID}" --zone "${ZONE}"

### Setting up Jenkins ###

The persistent disk should already be created. If not follow instruction from
[Google Cloud Plaform] (https://github.com/GoogleCloudPlatform/continuous-deployment-on-kubernetes).

In order to create the Jenkins Instance in the Kubernetes cluster

    $ kubectl apply -f k8s/jenkins/

Next we'll create an self signed SSL certificate that we'll feed to an ingress.

    $ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout /tmp/tls.key \
    -out /tmp/tls.crt -subj "/CN=jenkins/O=jenkins"

    $ kubectl apply -f k8s/jenkins/lb

    # Wait for the ingress to be available.
    $ kubectl describe ingress jenkins -n jenkins
    Name:                   jenkins
    Namespace:              jenkins
    Address:                35.186.199.210
    Default backend:        jenkins-ui:8080 (10.176.0.9:8080)
    TLS:
      tls terminates
    Rules:
      Host  Path    Backends
      ----  ----    --------
      *     *       jenkins-ui:8080 (10.176.0.9:8080)
    Annotations:
      https-target-proxy:           k8s-tps-jenkins-jenkins--86e70760fdd5a6e2
      static-ip:                    k8s-fw-jenkins-jenkins--86e70760fdd5a6e2
      target-proxy:                 k8s-tp-jenkins-jenkins--86e70760fdd5a6e2
      url-map:                      k8s-um-jenkins-jenkins--86e70760fdd5a6e2
      backends:                     {"k8s-be-30412--86e70760fdd5a6e2":"HEALTHY"}
      forwarding-rule:              k8s-fw-jenkins-jenkins--86e70760fdd5a6e2
      https-forwarding-rule:        k8s-fws-jenkins-jenkins--86e70760fdd5a6e2
    Events:
      FirstSeen     LastSeen        Count   From                            SubobjectPath   Type            Reason  Message
      ---------     --------        -----   ----                            -------------   --------        ------  -------
      36m           36m             1       {loadbalancer-controller }                      Normal          ADD     jenkins/jenkins
      35m           35m             1       {loadbalancer-controller }                      Normal          CREATE  ip: 35.186.199.210

In this example, Jenkins can be accessed at 35.186.199.210:80.

    $ export JENKINS_INSTANCE=35.186.199.210
    $ export JENKINS_PORT=80

#### Setup AppEngine Reverse Proxy ####

We need to limit authentication to our Jenkins. To do so we'll create a reverse
proxy using AppEngine and nginx. All we need to do checkout some code and deploy
to AppEngine.

    $ cd appengine-proxy
    $ sh setup.sh -a "${JENKINS_INSTANCE}" -b "${JENKINS_PORT}" -c "${PROJECT_ID}"

#### Jenkins Environment variables ####

Point your browser to https://istio-testing.appspot.com/configure.

In the Global Properties section, add the following environment variables:


    name: BAZEL_ARGS
    value: --batch --host_jvm_args=-Dbazel.DigestFunction=SHA1

    name: BAZEL_BUILD_ARGS
    value: --hazelcast_node=hazelcast.hazelcast.svc.cluster.local --spawn_strategy=remote

    name: BUCKET
    value: istio-artifacts

    name: GKE_CLUSTER
    value: jenkins-cluster

    name: PATH
    value: /usr/lib/google-cloud-sdk/bin:/usr/local/go/bin:${PATH}

    name: ZONE
    value: us-central1-f


### Updating Jenkins to point to the cluster ###

Point your browser to to https://istio-testing.appspot-preview.com/configure
and find the 'Kubernetes' section at the end of the page. The only things that
needs to be updated is 'Kubernetes URL' which should point to the URL above.

In https://istio-testing.appspot.com/configure, in the Global properties
section, make sure GKE_CLUSTER points to the value of ${CLUSTER_NAME}.

Other configuration details:

    # Cloud.Kubernetes Section
    Name: Jenkins Cluster
    Kubernetes URL: https://kubernetes.default
    Kubernetes Namespace: jenkins

    # Jenkins URL is jenkins local ip in pantheon plus ${JENKINS_PORT}
    Jenkins URL: http://jenkins-ui.jenkins.svc.cluster.local:8080
    Jenkins tunnel: jenkins-discovery.jenkins.svc.cluster.local:50000

    Container Cap: 100

    # Add 3 pod Templates:
    # Pod 1
    Name: ubuntu-16-04
    Labels: ubuntu-16-04
    Container Name: jnlp
    Docker iamge: gcr.io/istio-testing/ubuntu-16-04-slave:latest
    Always pull image: true
    Working Directory: /home/jenkins
    Command To run slave Agent:
    Arguments to pass to the command: ${computer.jnlpmac} ${computer.name}
    Max number of instances: 30
    Advanded:
      Run in privileged mode: true
      Request CPU: 500m
      Request Memory: 512Mi
      Limit CPU: 2000m
      Limit Memory: 8Gi
    # Pod 2
    Name: ubuntu-16-04-test
    Labels: ubuntu-16-04-test
    Container Name: jnlp
    Docker iamge: gcr.io/istio-testing/ubuntu-16-04-slave:test
    Always pull image: true
    Working Directory: /home/jenkins
    Command To run slave Agent:
    Arguments to pass to the command: ${computer.jnlpmac} ${computer.name}
    Max number of instances: 10
    Advanced:
      Run in privileged mode: true
      Request CPU: 500m
      Request Memory: 512Mi
      Limit CPU: 2000m
      Limit Memory: 4Gi
    # Pod 3
    Name: ubuntu-16-04-build
    Labels: ubuntu-16-04-build
    Container Name: jnlp
    Docker iamge: gcr.io/istio-testing/ubuntu-16-04-slave:latest
    Always pull image: true
    Working Directory: /home/jenkins
    Command To run slave Agent:
    Arguments to pass to the command: ${computer.jnlpmac} ${computer.name}
    Max number of instances: 10
    Advanced:
      Run in privileged mode: true
      Request CPU: 500m
      Request Memory: 512Mi
      Limit CPU: 4000m
      Limit Memory: 16Gi


### Creating an Hazelcast cluster ###

Hazelcast is used to create a distributed cache to Bazel, enabling faster
builds.

Let's build the docker image first

    $ cd k8s/hazelcast
    $ docker build -t gcr.io/istio-testing/hazelcast .
    $ gcloud docker push gcr.io/istio-testing/hazelcast

Let's deploy it:

    $ kubectl create -f hazelcast/deployment.yaml
    # Let's wait until the service is created
    $ kubectl describe service hazelcast --namespace hazelcast
    Name:                   hazelcast
    Namespace:              hazelcast
    Labels:                 name=hazelcast
    Selector:               name=hazelcast
    Type:                   ClusterIP
    IP:                     10.39.248.208
    Port:                   <unset> 5701/TCP
    Endpoints:              10.36.0.30:5701
    Session Affinity:       None
    No events.


## Maintenance ##

### Keeping Jenkins Plugins up to date ###

Go to the [Plugin Manager](https://istio-testing.appspot.com/pluginManager),
select all plugins, click update and check the restart checkbox.

## Troubleshooting ##

### All Slaves are marked as offline ###

Make sure you get the right credentials to manage the gke cluster:

    $ gcloud container clusters get-credentials jenkins-cluster-tmp \
    --project "${PROJECT_ID}" --zone us-central1-f

First lets look at the slaves on Kubernetes

    $ kubectl get pods --namespace jenkins
    NAME                              READY     STATUS    RESTARTS   AGE
    debian-jessie-slave-1b74d70d416   1/1       Running   0          9m
    debian-jessie-slave-1b9a0a87d8c   1/1       Running   0          8m
    debian-jessie-slave-1bbf4b70753   1/1       Running   0          8m
    debian-jessie-slave-1be48e2bfdf   1/1       Running   0          8m
    debian-jessie-slave-1c2f10b2da0   1/1       Running   0          8m

if most listed pods are in Running or ContainerCreating status, then it must be
an issue with Jenkins. You should have the same number of running (as opposed to
offline or suspended) slave in Jenkins as the one list here. To find the list of
slaves point your browser to https://istio-testing.appspot.com/computer/.

In any case, while we cleanup we want to stop new builds from starting. To do so
point your browser to https://istio-testing.appspot.com/quietDown

Now that no new build should start, we need to stop all running builds. You
might have to go to the console and click on printed links to forcibly kill the
build. Once all builds are killed, you need to delete each slave, but clicking
on them, selecting Delete Slave and confirming.

In order to understand if there is something wrong with the kubernetes cluster,
check all the pods:

    $ kubectl get pods --all-namespaces

If all pods are not in Running status, then there is something wrong with the
cluster. It would be good to have someone in the Kubernetes team to take a look
but the easy thing is to go to delete the cluster

    $ gcloud container clusters delete "${CLUSTER_NAME}" \
    --project "${PROJECT_ID}" \
    --zone "${ZONE}"

Next go the deployment section and recreate the cluster, jenkins and hazelcast.
You will also need to redeploy the app-engine proxy.


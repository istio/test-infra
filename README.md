## Table of Content

- [Deployment information](#deployment-information)
- [Kubernetes Setup](#kubernetes-setup)
  * [Creating Jenkins Cluster](#creating-jenkins-cluster)
  * [Create a persistent disk Jenkins home](#create-a-persistent-disk-jenkins-home)
  * [Setting up Jenkins](#setting-up-jenkins)
    + [Setup AppEngine Reverse Proxy](#setup-appengine-reverse-proxy)
  * [Creating an Hazelcast cluster](#creating-an-hazelcast-cluster)
- [Maintenance](#maintenance)
  * [Preparing for update](#preparing-for-update)
  * [Keeping Jenkins Setup up to date](#keeping-jenkins-setup-up-to-date)
  * [Upgrading Jenkins plugins](#upgrading-jenkins-plugins)
  * [Reverting to a working configuration](#reverting-to-a-working-configuration)
- [Troubleshooting](#troubleshooting)
  * [All Slaves are marked as offline](#all-slaves-are-marked-as-offline)

## Deployment information ##

    $ export PROJECT_ID='istio-testing'
    $ export ZONE='us-central1-f'
    $ export CLUSTER_NAME='jenkins-cluster'
    $ export K8S_SCOPES='https://www.googleapis.com/auth/appengine.admin,https://www.googleapis.com/auth/cloud-platform,https://www.googleapis.com/auth/compute,https://www.googleapis.com/auth/devstorage.full_control,https://www.googleapis.com/auth/logging.write,https://www.googleapis.com/auth/projecthosting,https://www.googleapis.com/auth/servicecontrol,https://www.googleapis.com/auth/service.management'

## Kubernetes Setup ##

### Creating Jenkins Cluster ###

Jenkins runs all its slaves in a Kubernetes Cluster. The cluster name is
hardcoded in the Jenkinsfile.

    $ gcloud container clusters create "${CLUSTER_NAME}" \
    --project="${PROJECT_ID}" \
    --zone="${ZONE}" \
    --machine-type=n1-standard-8 \
    --disk-size=100 \
    --scopes="${K8S_SCOPES}" \
    --num-nodes=8 \
    --image-type=gci \
    --network=default \
    --enable-cloud-logging \
    --no-enable-cloud-monitoring \
    --no-enable-cloud-endpoints \
    --local-ssd-count=1 \
    --node-labels=role=build

    # Update kubectl config.
    # You might want to run this command as well on your desktop.
    $ gcloud container clusters get-credentials "${CLUSTER_NAME}" \
    --project "${PROJECT_ID}" --zone "${ZONE}"

### Create a persistent disk Jenkins home ###

This step should only be necessary if a disk does not already exist.

If there is no existing persistent disk for Jenkins, we need to create one from
source code. To do so, let's use the script/create_backup_pd script

    $ scripts/create_backup_pd -h
    Usage:
        -z  Specify zone (Optional, default:us-central1-f)
        -b  Bucket name where backup disk be stored (Optional,
            default:istio-tools/jenkins/images)
        -s  SHA which collected jenkins secrets (Necessry)
        -t  Bucket where secrets are stored (Optional,
            default:istio-tools/jenkins-secrets)
        -d  Name for backup disk (Optional, default:jenkins-home)


The only required argument is the SHA at which the backup was created. You can find those by running

     $ git log jenkins_home/


For this example we'll use 1623630acb3d3a1b79d687647f162e7f76501e2a and create a disk named jenkins-home-backup.

    $ export SHA='1623630acb3d3a1b79d687647f162e7f76501e2a'
    $ ./create_backup_pd -s ${SHA} \
        -d jenkins-home-backup

Disk created via this method do not use partition, so we need to update
k8s/jenkins/jenkins.yaml to use the raw device directly and use the disk
that we just created

From

    pdName: jenkins
    fsType: ext4
    partition: 1

To:

    pdName: jenkins-home-backup.
    fsType: ext4
    partition: 0

### Setting up Jenkins ###

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
    $ ./setup.sh -a "${JENKINS_INSTANCE}" -b "${JENKINS_PORT}" -c "${PROJECT_ID}"

### Creating an Hazelcast cluster ###

This is not required anymore. You can skip the next section.
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

### Preparing for update ###

When doing an upgrade, it is better to make sure that no jobs are running. In
order to prevent new test from running, direct your browser to
https://istio-testing.appspot.com/quietDown.

Checkout the last version of istio-testing.

Next run this script, which will backup important jenkins persistent data and
make a snapshot of the Persistent Disk used for Jenkins.

    $ scripts/prepare_for_upgrade

This script will create a commit to your local checkout. Please create a PR and
have the PR be merged in.

### Keeping Jenkins Setup up to date ###

For jenkins upgrade, we just need to update the k8s/jenkins/jenkins.yaml file
and apply the configuration, check that everything works as expected and save the change to
source code.

Update k8s/jenkins/jenkins.yaml from:

    spec:
      containers:
      - name: master
        image: jenkins:2.32.3

to the new version of Jenkinsi (Here 2.32.4 as an example):

    spec:
      containers:
      - name: master
        image: jenkins:2.32.4

And apply the changes:

    kubectl apply -f k8s/jenkins/jenkins.yaml


Please don't forget to commit and push your changes.


### Upgrading Jenkins plugins ###

Jenkins plugins upgrade needs to be done via the UI. It is recommended
to check the each plugin changelog to understand what the impact could be.
The plugins that more likely to break workflows are the one related to
github, pipeline, and kubernetes-plugin.

Go to the [Plugin Manager](https://istio-testing.appspot.com/pluginManager),
select all plugins, click update and check the restart checkbox.


### Reverting to a working configuration ###

Check the existing disk snapshots

    $ gcloud compute snapshots list
    NAME                   DISK_SIZE_GB  SRC_DISK                     STATUS
    jenkins-03-08-17-1204  100           us-central1-f/disks/jenkins  READY
    jenkins-03-08-17-1417  100           us-central1-f/disks/jenkins  READY

And create a disk from the snapshot you created before the upgrade:

    $ gcloud compute disks create jenkins-home \
       --type "pd-ssd" \
       --source-snapshot=jenkins-03-08-17-1417 \
       --project ${PROJECT_ID} \
       --zone ${ZONE}

Update k8s/jenkins/jenkins.yaml from:

    volumes:
      - name: jenkins-home
        gcePersistentDisk:
          pdName: jenkins
          fsType: ext4
          partition: 1

with the new created disk:

    volumes:
      - name: jenkins-home
        gcePersistentDisk:
          pdName: jenkins-home
          fsType: ext4
          partition: 1

And apply the changes:

    kubectl apply -f k8s/jenkins/jenkins.yaml

Please don't forget to commit and push your changes.

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


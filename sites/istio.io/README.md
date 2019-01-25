Overview
====
This contains the Nginx configuration for istio.io and the associated subdomain
redirectors.

Testing
===
Configure kubectl to target a test cluster on GKE.

Run `make deploy-fake-secret deploy` and wait for the service to be available--
the load balancer may take some time to configure.

Set `TARGET_IP` to the ingress IP of the running service:

    export TARGET_IP=$(kubectl get svc istio-io '--template={{range .status.loadBalancer.ingress}}{{.ip}}{{end}}')

Use `make test` to run unit tests to verify the various endpoints on the server.

Deploying
===

For the deployment set kubectl to target the production cluster
and run `make deploy`. Nginx doesn't auto-detect configuration file
changes so pods may need to be manually killed to force deployment to
restart nginx with new configuration via ConfigMaps.

Design
===

This repo is based on https://github.com/kubernetes/k8s.io and uses
various Google Cloud platform services for the actual
infrastructure. This document describes the high-level design for the
istio.io infrastructure beyond what is already documented by the
kubernetes YAML files. When in doubt read the source and
https://cloud.google.com/dns/docs or open an GH issue. There are
certainly ways to improve upon this approach so suggestions and PR
welcome.

The basic idea is to use a L4 load balancer in front of a cluster of
nginx proxies. The L4 load balancer distributes traffic to all healthy
proxy instances whose number can be scaled up/down based on system
load and desired reliability.

### DNS and addresses

Cloud DNS is used for easy automatic management of DNS records,
e.g. auto-certificate renewal via DNS challenge with ACMA client can
be easily integrated into cluster.

    $ gcloud dns record-sets list --zone=istio-io
    NAME                               TYPE   TTL    DATA
    istio.io.                          A      300    35.185.199.142
    istio.io.                          NS     21600  ns-cloud-e1.googledomains.com.,ns-cloud-e2.googledomains.com.,ns-cloud-e3.googledomains.com.,ns-cloud-e4.googledomains.com.
    istio.io.                          SOA    21600  ns-cloud-e1.googledomains.com. cloud-dns-hostmaster.google.com. 1 21600 3600 259200 300
    www.istio.io.                      CNAME  300    istio.io.


    $ gcloud compute addresses list
    NAME             REGION    ADDRESS         STATUS
    istio-io         us-west1  35.233.169.38   IN_USE

### Network load balancer and kubernetes services

A kubernetes service with `type: LoadBalancer` maps an externally
accessible IP to the backend services. GKE does most of the work here
and sets up the necessary GCP load balancer rules when the kubernetes
service is created.

- [`service.yaml`](https://github.com/istio/istio.io/test-infra/sites/istio.io/service.yaml) handles *.istio.io.

### TLS

Certificates for TLS termination are from [Let's
Encrypt](https://letsencrypt.org/). The current certificates are
generated out-of-band and provided to nginx proxies via kubernetes
secrets. Something like
[kube-cert-manager](https://www.google.com/webhp?sourceid=chrome-instant&ion=1&espv=2&ie=UTF-8#q=kube-cert-manager&*)
could be used to automate certificate renewal.

- Manual method:

Use [certbot](https://certbot.eff.org/) to renew certificates for the
necessary (sub)domains. Add the required DNS `TXT` records through
Cloud DNS. Its best to wait 5-10 minutes after DNS records are updated
before completing verification as it may take some time for DNS
changes to propogate and be picked up by Let's Encrypt's verification
server.

    sudo certbot-auto certonly --manual -d istio.io,www.istio.io,velodrome.istio.io,gcsweb.istio.io,prow.istio.io --preferred-challenges=dns

Backup current certificate secret in case changes need to be rolled
back.

    kubectl get secret istio.io -o yaml > previous-secret-istio-io.yaml

Copy generated key and certificate file to current directory and
refresh kubernetes certificate secret.

    sudo cp /etc/letsencrypt/live/istio.io/privkey.pem tls.key
    sudo cp /etc/letsencrypt/live/istio.io/fullchain.pem tls.crt
    kubectl create secret generic istio.io --from-file=tls.key --from-file=tls.crt --dry-run -o yaml | kubectl apply -f -

Force nginx to pick-up the new certificates by using `/usr/sbin/nginx -s reload`
or delete the nginx pods and letting kubernetes restart with udpated
configuration and certs.

- Automatic method: *TODO*

Download trusted CA certificates for backend (e.g. istio.github.io)
verification and store as a kubernetes secret.

    curl -O https://www.digicert.com/CACerts/DigiCertHighAssuranceEVRootCA.crt
    curl -O https://www.geotrust.com/resources/root_certificates/certificates/GeoTrust_Global_CA.pem

    openssl x509 -inform DER -in DigiCertHighAssuranceEVRootCA.crt -outform PEM -out DigiCertHighAssuranceEVRootCA.pem

    kubectl create secret generic \
        --from-file=DigiCertHighAssuranceEVRootCA.pem \
        --from-file=GeoTrust_Global_CA.pem \
        --dry-run -o yaml > secret-cacerts.yaml

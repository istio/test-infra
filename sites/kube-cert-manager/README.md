# Creating a HTTPs Nginx ingress

We are using cert-manager with Let's encrypt

More detail information at <https://cert-manager.readthedocs.io/en/latest/getting-started/1-configuring-helm.html>

## Installing Cert Manager

Make sure you have helm installed.

Create a ClusterRoleBinding such that tiller can can install deployments in
every namespace:

    kubectl apply -f prow/cluster/rbac-config.yaml

Then install tiller

    helm init --service-account tiller

Once helm is installed properly, you can proceed by installing cert-manager

    helm install \
    --name cert-manager \
    --namespace kube-system \
    stable/cert-manager

## Setting Up Cert Management

We are using a cluster issuer to generate certs with Let's encrypt using DNS
validation

    kubectl apply -f prow/cluster/cluster-issuer.yaml

We can now have create a cert

    kubectl apply -f istio.io-cert.yaml

This will automatically create the istio-io-tls secret containing the certs used
in the nginx deployment.

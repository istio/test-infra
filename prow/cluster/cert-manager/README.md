# Creating a HTTPs ingress

We are using cert-manager with Let's encrypt

More detail information at <https://cert-manager.readthedocs.io/en/latest/getting-started/1-configuring-helm.html>

## Installing Cert Manager

Make sure you have helm installed.

Create a ClusterRoleBinding such that tiller can can install deployments in
every namespace:

    kubectl apply -f rbac-config.yaml

Then install tiller

    helm init --service-account tiller

Once helm is installed properly, you can proceed by installing cert-manager

    helm install \
    --name cert-manager \
    --namespace kube-system \
    stable/cert-manager

## Setting Up Ingress

We are using a cluster issuer to generate certs with Let's encrypt

    kubectl apply -f cluster-issuer.yaml

Now we can setup the ingress

    kubectl apply -f tls-ing.yaml

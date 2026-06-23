# Cluster Autoscaler. One deployment per cluster; each one scales only the
# managed node groups of its own cluster via ASG auto-discovery. EKS managed
# node groups automatically tag their Auto Scaling Groups with
# k8s.io/cluster-autoscaler/enabled and k8s.io/cluster-autoscaler/<cluster>, so
# no extra tagging is required for discovery.
#
# Each controller's IAM role is delivered via EKS Pod Identity: the association
# binds the role to the kube-system/cluster-autoscaler service account, scoped
# to that cluster's ASGs only.

locals {
  cluster_autoscaler_chart_version = "9.46.6"
  cluster_autoscaler_image_tag     = "v${local.cluster_version}.0"
}

module "cluster_autoscaler_identity" {
  source  = "terraform-aws-modules/eks-pod-identity/aws"
  version = "~> 2.0"

  for_each = local.clusters

  name = "cluster-autoscaler-${each.key}"

  attach_cluster_autoscaler_policy = true
  cluster_autoscaler_cluster_names = [module.eks[each.key].cluster_name]

  associations = {
    this = {
      cluster_name    = module.eks[each.key].cluster_name
      namespace       = "kube-system"
      service_account = "cluster-autoscaler"
    }
  }
}

resource "helm_release" "cluster_autoscaler_prow" {
  name       = "cluster-autoscaler"
  repository = "https://kubernetes.github.io/autoscaler"
  chart      = "cluster-autoscaler"
  version    = local.cluster_autoscaler_chart_version
  namespace  = "kube-system"

  set {
    name  = "autoDiscovery.clusterName"
    value = module.eks["prow"].cluster_name
  }

  set {
    name  = "awsRegion"
    value = local.region
  }

  set {
    name  = "image.tag"
    value = local.cluster_autoscaler_image_tag
  }

  set {
    name  = "rbac.serviceAccount.create"
    value = "true"
  }

  set {
    name  = "rbac.serviceAccount.name"
    value = "cluster-autoscaler"
  }

  depends_on = [module.cluster_autoscaler_identity]
}

resource "helm_release" "cluster_autoscaler_prow_build" {
  provider = helm.prow_build

  name       = "cluster-autoscaler"
  repository = "https://kubernetes.github.io/autoscaler"
  chart      = "cluster-autoscaler"
  version    = local.cluster_autoscaler_chart_version
  namespace  = "kube-system"

  set {
    name  = "autoDiscovery.clusterName"
    value = module.eks["prow-build"].cluster_name
  }

  set {
    name  = "awsRegion"
    value = local.region
  }

  set {
    name  = "image.tag"
    value = local.cluster_autoscaler_image_tag
  }

  set {
    name  = "rbac.serviceAccount.create"
    value = "true"
  }

  set {
    name  = "rbac.serviceAccount.name"
    value = "cluster-autoscaler"
  }

  depends_on = [module.cluster_autoscaler_identity]
}

resource "helm_release" "cluster_autoscaler_prow_private" {
  provider = helm.prow_private

  name       = "cluster-autoscaler"
  repository = "https://kubernetes.github.io/autoscaler"
  chart      = "cluster-autoscaler"
  version    = local.cluster_autoscaler_chart_version
  namespace  = "kube-system"

  set {
    name  = "autoDiscovery.clusterName"
    value = module.eks["prow-private"].cluster_name
  }

  set {
    name  = "awsRegion"
    value = local.region
  }

  set {
    name  = "image.tag"
    value = local.cluster_autoscaler_image_tag
  }

  set {
    name  = "rbac.serviceAccount.create"
    value = "true"
  }

  set {
    name  = "rbac.serviceAccount.name"
    value = "cluster-autoscaler"
  }

  depends_on = [module.cluster_autoscaler_identity]
}

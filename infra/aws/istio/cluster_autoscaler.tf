locals {
  cluster_autoscaler_chart_version = "9.58.0"
  cluster_autoscaler_image_tag     = "v1.35.0"
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

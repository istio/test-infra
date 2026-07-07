# Cross-cluster access for the Prow control plane.
#
# The control-plane components run on the `prow` cluster but talk to the build
# clusters' API servers (via the aws-iam-authenticator exec plugin bundled in
# the prow images; see prow/aws/cluster/build_kubeconfig.yaml) to schedule,
# reap, and read job pods. Each component needs two things:
#
# Writers (prow-controller-manager, sinker) create and delete pods -> Edit.
# Readers (deck, deck-private, crier) only read pod logs/status -> View, reusing

locals {
  # Components that create and delete job pods on the build clusters. They get
  # dedicated IAM roles with no AWS API permissions; the role exists solely as
  # an identity for aws-iam-authenticator.
  prow_writer_components = {
    prow-controller-manager = { service_account = "prow-controller-manager" }
    sinker                  = { service_account = "sinker" }
  }

  # Clusters whose API servers the control plane reaches as build clusters.
  prow_build_clusters = ["prow", "prow-build", "prow-private"]

  # Read-only components reuse their existing workload roles and only need to
  # reach the clusters that actually run the pods they surface.
  prow_reader_components = {
    deck         = { clusters = ["prow-build", "prow"], role_arn = module.workload_identity["deck"].iam_role_arn }
    deck-private = { clusters = ["prow-private"], role_arn = module.workload_identity["deck-private"].iam_role_arn }
    crier        = { clusters = ["prow-build", "prow-private", "prow"], role_arn = module.workload_identity["crier"].iam_role_arn }
  }

  prow_writer_access = merge([
    for name, cfg in local.prow_writer_components : {
      for cluster in local.prow_build_clusters :
      "${name}.${cluster}" => {
        cluster   = cluster
        principal = module.prow_control_plane_identity[name].iam_role_arn
        policy    = "AmazonEKSEditPolicy"
      }
    }
  ]...)

  prow_reader_access = merge([
    for name, cfg in local.prow_reader_components : {
      for cluster in cfg.clusters :
      "${name}.${cluster}" => {
        cluster   = cluster
        principal = cfg.role_arn
        policy    = "AmazonEKSViewPolicy"
      }
    }
  ]...)

  prow_cluster_access = merge(local.prow_writer_access, local.prow_reader_access)
}

module "prow_control_plane_identity" {
  source  = "terraform-aws-modules/eks-pod-identity/aws"
  version = "~> 2.0"

  for_each = local.prow_writer_components

  name = each.key

  # No AWS API permissions: the role is only an identity for cross-cluster
  # authentication, not for reaching any AWS service.
  attach_custom_policy = false

  associations = {
    prow = {
      cluster_name    = module.eks["prow"].cluster_name
      namespace       = "default"
      service_account = each.value.service_account
    }
  }
}

resource "aws_eks_access_entry" "prow_control_plane" {
  for_each = local.prow_cluster_access

  cluster_name  = module.eks[each.value.cluster].cluster_name
  principal_arn = each.value.principal
}

resource "aws_eks_access_policy_association" "prow_control_plane" {
  for_each = local.prow_cluster_access

  cluster_name  = module.eks[each.value.cluster].cluster_name
  principal_arn = each.value.principal
  policy_arn    = "arn:aws:eks::aws:cluster-access-policy/${each.value.policy}"

  access_scope {
    type       = "namespace"
    namespaces = ["test-pods"]
  }

  depends_on = [aws_eks_access_entry.prow_control_plane]
}

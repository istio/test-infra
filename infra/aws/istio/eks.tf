# eks.tf provisions the EKS clusters. These are skeletons: minimal node groups
# just to stand the clusters up. Pools get sized to match the GKE workloads in a
# later pass.
#
# Cluster mapping from GCP:
#   prow         <- GKE `prow` in istio-testing      (control plane / trusted)
#   prow-build   <- GKE `prow` in istio-prow-build    (core build)
#                   + GKE `prow-arm` (merged: GCP needed a separate ARM cluster
#                     because Graviton equivalents were single-zone; on AWS ARM
#                     is multi-AZ, so it's just another node group here)
#   prow-private <- GKE `prow` in istio-prow-private  (private / PSWG build)

locals {
  # EKS control-plane version. Bump deliberately; node groups follow.
  cluster_version = "1.33"

  clusters = {
    prow = {
      node_groups = {
        default = {
          ami_type       = "AL2023_x86_64_STANDARD"
          instance_types = ["t3.small"]
          min_size       = 1
          max_size       = 3
          desired_size   = 1
        }
      }
    }

    prow-build = {
      node_groups = {
        x86 = {
          ami_type       = "AL2023_x86_64_STANDARD"
          instance_types = ["t3.small"]
          min_size       = 1
          max_size       = 3
          desired_size   = 1
        }
        # Graviton pool — replaces the standalone GKE `prow-arm` cluster.
        arm = {
          ami_type       = "AL2023_ARM_64_STANDARD"
          instance_types = ["t4g.small"]
          min_size       = 1
          max_size       = 3
          desired_size   = 1
        }
      }
    }

    prow-private = {
      node_groups = {
        default = {
          ami_type       = "AL2023_x86_64_STANDARD"
          instance_types = ["t3.small"]
          min_size       = 1
          max_size       = 3
          desired_size   = 1
        }
      }
    }
  }
}

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 21.0"

  for_each = local.clusters

  name               = each.key
  kubernetes_version = local.cluster_version

  vpc_id     = module.vpc[each.key].vpc_id
  subnet_ids = module.vpc[each.key].private_subnets

  # Auth via EKS access entries (API). This is the boundary mechanism that
  # replaces GCP project isolation: each principal is granted RBAC only on the
  # clusters it may use. The Terraform principal gets admin to bootstrap.
  authentication_mode                      = "API"
  enable_cluster_creator_admin_permissions = true

  # Skeleton: public endpoint so the API is reachable while building out.
  # Lock this down (private endpoint + CIDR allow-list) before production.
  endpoint_public_access = true

  addons = {
    coredns    = {}
    kube-proxy = {}
    vpc-cni    = {}
    # Required for the workload IAM roles in iam.tf (EKS Pod Identity).
    eks-pod-identity-agent = {}
  }

  eks_managed_node_groups = each.value.node_groups
}

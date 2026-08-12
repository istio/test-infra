locals {
  cluster_version = "1.36"
  gp3_root_volume = {
    delete_on_termination = true
    encrypted             = true
    volume_type           = "gp3"
  }

  clusters = {
    prow = {
      node_groups = {
        default = {
          ami_type       = "AL2023_x86_64_STANDARD"
          instance_types = ["t3.medium"]
          capacity_type  = "ON_DEMAND"
          min_size       = 1
          max_size       = 5
          desired_size   = 3
          block_device_mappings = {
            root = {
              device_name = "/dev/xvda"
              ebs         = merge(local.gp3_root_volume, { volume_size = 100 })
            }
          }
          labels = { prod = "prow" }
        }
      }
    }

    # Core build cluster (for both amd and arm)
    prow-build = {
      node_groups = {
        # Large build pool
        build = {
          ami_type       = "AL2023_x86_64_STANDARD"
          instance_types = ["m6i.16xlarge"]
          capacity_type  = "ON_DEMAND"
          min_size       = 0
          max_size       = 5
          desired_size   = 1
          block_device_mappings = {
            root = {
              device_name = "/dev/xvda"
              ebs         = merge(local.gp3_root_volume, { volume_size = 2000 })
            }
          }
          labels = { testing = "build-pool" }
        }
        # Primary test pool
        test-e2 = {
          ami_type       = "AL2023_x86_64_STANDARD"
          instance_types = ["m6i.4xlarge"]
          capacity_type  = "ON_DEMAND"
          min_size       = 1
          max_size       = 5
          desired_size   = 1
          block_device_mappings = {
            root = {
              device_name = "/dev/xvda"
              ebs         = merge(local.gp3_root_volume, { volume_size = 256 })
            }
          }
          labels = { testing = "test-pool" }
        }
        arm = {
          ami_type       = "AL2023_ARM_64_STANDARD"
          instance_types = ["m7g.4xlarge"]
          capacity_type  = "ON_DEMAND"
          min_size       = 0
          max_size       = 5
          desired_size   = 1
          block_device_mappings = {
            root = {
              device_name = "/dev/xvda"
              ebs         = merge(local.gp3_root_volume, { volume_size = 256 })
            }
          }
          labels = { testing = "test-pool" }
        }
      }
    }

    prow-private = {
      node_groups = {
        # Test pool (e2-standard-16).
        test = {
          ami_type       = "AL2023_x86_64_STANDARD"
          instance_types = ["t3.large"]
          capacity_type  = "ON_DEMAND"
          min_size       = 1
          max_size       = 5
          desired_size   = 1
          block_device_mappings = {
            root = {
              device_name = "/dev/xvda"
              ebs         = merge(local.gp3_root_volume, { volume_size = 256 })
            }
          }
          labels = { testing = "test-pool" }
        }
        # High-memory build pool
        build = {
          ami_type       = "AL2023_x86_64_STANDARD"
          instance_types = ["t3.small"]
          capacity_type  = "ON_DEMAND"
          min_size       = 0
          max_size       = 5
          desired_size   = 1
          block_device_mappings = {
            root = {
              device_name = "/dev/xvda"
              ebs         = merge(local.gp3_root_volume, { volume_size = 2000 })
            }
          }
          labels = { testing = "build-pool" }
        }
        arm = {
          ami_type       = "AL2023_ARM_64_STANDARD"
          instance_types = ["t4g.small"]
          capacity_type  = "ON_DEMAND"
          min_size       = 0
          max_size       = 5
          desired_size   = 1
          block_device_mappings = {
            root = {
              device_name = "/dev/xvda"
              ebs         = merge(local.gp3_root_volume, { volume_size = 256 })
            }
          }
          labels = { testing = "test-pool" }
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

  addons = merge({
    coredns = each.key == "prow" ? {
      configuration_values = jsonencode({
        podAnnotations = {
          "cluster-autoscaler.kubernetes.io/safe-to-evict" = "true"
        }
      })
      } : each.key == "prow-build" ? {
      configuration_values = jsonencode({
        nodeSelector = {
          testing = "test-pool"
        }
      })
    } : {}
    # CNI and kube-proxy must be installed before the node groups so nodes can
    # get pod networking and reach Ready; otherwise node group creation deadlocks
    # waiting on nodes that can never join.
    kube-proxy = {
      before_compute = true
    }
    vpc-cni = {
      before_compute = true
    }
    # Required for the workload IAM roles in iam.tf (EKS Pod Identity).
    eks-pod-identity-agent = {}
    }, each.key == "prow" ? {
    aws-ebs-csi-driver = {
      configuration_values = jsonencode({
        controller = {
          # cluster autoscaler is wary about evicting this because it has an emptydir
          podAnnotations = {
            "cluster-autoscaler.kubernetes.io/safe-to-evict" = "true"
          }
        }
      })
      pod_identity_association = [{
        role_arn        = module.ebs_csi_identity.iam_role_arn
        service_account = "ebs-csi-controller-sa"
      }]
    }
  } : {})

  eks_managed_node_groups = each.value.node_groups
}

# iam.tf defines the IAM roles assumed by in-cluster workloads via EKS Pod
# Identity. Each role mirrors a GCP service account / Workload Identity binding.
#
# Because object storage moved to Cloudflare R2 and registries to GHCR, most
# workloads' cloud permissions reduce to "read (and sometimes write) specific
# Secrets Manager secrets" (the R2 credentials / tokens). GCS/GCR/RBE grants
# from the GCP config have no AWS equivalent and are intentionally dropped.
#
# The community eks-pod-identity module builds the role + Pod Identity trust
# policy (pods.eks.amazonaws.com) and our scoped secrets policy. Pod Identity
# *associations* (binding a role to a cluster + namespace + service account) are
# wired in the identity-secrets phase; see the example at the bottom.

locals {
  # Workloads that need an AWS IAM role. `read` / `write` reference secret keys
  # in aws_secretsmanager_secret.secrets (see secrets.tf).
  #
  # Excluded by design:
  #   - prowjob-rbe:             RBE is GCP-specific; no AWS equivalent.
  #   - opentelemetry-collector: Cloud Trace -> X-Ray; different perm model.
  #   - prow-deployer / prow-control-plane: cluster access is granted via EKS
  #                              access entries, not IAM policies (EKS phase).
  #   - istio-policy-bot / prowjob-bots-deployer: policy bot is a later phase.
  workload_roles = {
    # Highly privileged release job. Also needs KMS (cosign) access, added with
    # the KMS key in a later phase.
    "prowjob-release" = {
      read = [
        "release_docker_istio",
        "release_github_istio-release",
        "release_grafana_istio",
        "github_istio-testing_pusher",
        "cf_r2_istio-prerelease_credentials",
        "cf_r2_istio-release_credentials",
      ]
      associations = {
        prow-build = { namespace = "test-pods", service_account = "prowjob-release" }
      }
    }

    "prowjob-github-read" = {
      read         = ["github-read_github_read"]
      associations = { prow-build = { namespace = "test-pods", service_account = "prowjob-github-read" } }
    }
    "prowjob-github-istio-testing" = {
      read         = ["github_istio-testing_pusher"]
      associations = { prow-build = { namespace = "test-pods", service_account = "prowjob-github-istio-testing" } }
    }
    "prowjob-build-tools" = {
      read         = ["github_istio-testing_pusher"]
      associations = { prow-build = { namespace = "test-pods", service_account = "prowjob-build-tools" } }
    }
    # Runs on both the build cluster and the trusted control-plane cluster.
    "prowjob-testing-write" = {
      read = ["cf_r2_istio-build_credentials"]
      associations = {
        prow-build = { namespace = "test-pods", service_account = "prowjob-testing-write" }
        prow       = { namespace = "test-pods", service_account = "prowjob-testing-write" }
      }
    }

    # ESO (External Secrets Operator) controller identity is defined per cluster
    # in external_secrets.tf, not here.

    # Default Prow job service account.
    "prowjob-default" = {
      read = [
        "cf_r2_istio-prow_credentials",
        "cf_r2_public_buckets_ro_credentials",
      ]
      associations = { prow-build = { namespace = "test-pods", service_account = "prowjob-default-sa" } }
    }

    # Rotates ephemeral Cloudflare R2 credentials: reads the permanent admin
    # token + all per-bucket creds, and writes new versions of the per-bucket
    # creds. Runs as a CronJob in the trusted control-plane cluster.
    "cloudflare-rotator" = {
      read = [
        "cf_r2_admin_token",
        "cf_r2_istio-build_credentials",
        "cf_r2_istio-build-private_credentials",
        "cf_r2_istio-prerelease_credentials",
        "cf_r2_istio-prerelease-private_credentials",
        "cf_r2_istio-prow_credentials",
        "cf_r2_istio-prow-private_credentials",
        "cf_r2_istio-testgrid_credentials",
      ]
      write = [
        "cf_r2_istio-build_credentials",
        "cf_r2_istio-build-private_credentials",
        "cf_r2_istio-prerelease_credentials",
        "cf_r2_istio-prerelease-private_credentials",
        "cf_r2_istio-prow_credentials",
        "cf_r2_istio-prow-private_credentials",
        "cf_r2_istio-testgrid_credentials",
      ]
      associations = { prow = { namespace = "cloudflare-secret-rotation", service_account = "cloudflare-rotator" } }
    }
  }
}

module "workload_identity" {
  source  = "terraform-aws-modules/eks-pod-identity/aws"
  version = "~> 2.0"

  for_each = local.workload_roles

  name = each.key

  attach_custom_policy = true
  policy_statements = concat(
    [{
      sid       = "ReadSecrets"
      effect    = "Allow"
      actions   = ["secretsmanager:GetSecretValue", "secretsmanager:DescribeSecret"]
      resources = [for s in each.value.read : aws_secretsmanager_secret.secrets[s].arn]
    }],
    length(try(each.value.write, [])) > 0 ? [{
      sid       = "WriteSecrets"
      effect    = "Allow"
      actions   = ["secretsmanager:PutSecretValue"]
      resources = [for s in each.value.write : aws_secretsmanager_secret.secrets[s].arn]
    }] : [],
  )

  # Pod Identity associations bind this role to a cluster + namespace + k8s
  # service account. The map key references the EKS cluster (see eks.tf); the
  # cluster_name comes from that module so associations wait for the cluster.
  associations = {
    for cluster, assoc in each.value.associations : cluster => {
      cluster_name    = module.eks[cluster].cluster_name
      namespace       = assoc.namespace
      service_account = assoc.service_account
    }
  }
}

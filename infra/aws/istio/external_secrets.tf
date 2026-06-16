# External Secrets Operator (ESO). ESO syncs AWS Secrets Manager secrets into
# native Kubernetes Secrets, driven by ExternalSecret resources in cluster/eks/.
# It runs on every cluster that consumes secrets.
#
# Auth uses EKS Pod Identity: the controller's service account
# (external-secrets/external-secrets) is bound to a per-cluster IAM role scoped
# to exactly the secrets that cluster reads. A ClusterSecretStore with no auth
# block (see cluster/eks/.../external_secrets_store.yaml) therefore resolves
# against the controller pod's identity.

locals {
  # Secrets each cluster's ESO controller is allowed to read. Keys reference
  # aws_secretsmanager_secret.secrets (secrets.tf).
  eso_read_secrets = {
    prow = [
      "github_istio-testing_pusher",
      "cf_r2_istio-prow_credentials",
      "cf_r2_public_buckets_ro_credentials",
      "oauth_token",
      "istio-testing_robot-ssh-key",
    ]
    prow-build = [
      "cf_r2_istio-prow_credentials",
    ]
    prow-private = [
      "cf_r2_istio-prow-private_credentials",
    ]
  }
}

module "external_secrets_identity" {
  source  = "terraform-aws-modules/eks-pod-identity/aws"
  version = "~> 2.0"

  for_each = local.eso_read_secrets

  name = "external-secrets-${each.key}"

  attach_custom_policy = true
  policy_statements = [{
    sid       = "ReadSecrets"
    effect    = "Allow"
    actions   = ["secretsmanager:GetSecretValue", "secretsmanager:DescribeSecret"]
    resources = [for s in each.value : aws_secretsmanager_secret.secrets[s].arn]
  }]

  associations = {
    (each.key) = {
      cluster_name    = module.eks[each.key].cluster_name
      namespace       = "external-secrets"
      service_account = "external-secrets"
    }
  }
}

resource "helm_release" "external_secrets_prow" {
  name             = "external-secrets"
  repository       = "https://charts.external-secrets.io"
  chart            = "external-secrets"
  version          = "2.6.0"
  namespace        = "external-secrets"
  create_namespace = true

  set {
    name  = "installCRDs"
    value = "true"
  }

  depends_on = [module.external_secrets_identity]
}

resource "helm_release" "external_secrets_prow_build" {
  provider = helm.prow_build

  name             = "external-secrets"
  repository       = "https://charts.external-secrets.io"
  chart            = "external-secrets"
  version          = "2.6.0"
  namespace        = "external-secrets"
  create_namespace = true

  set {
    name  = "installCRDs"
    value = "true"
  }

  depends_on = [module.external_secrets_identity]
}

resource "helm_release" "external_secrets_prow_private" {
  provider = helm.prow_private

  name             = "external-secrets"
  repository       = "https://charts.external-secrets.io"
  chart            = "external-secrets"
  version          = "2.6.0"
  namespace        = "external-secrets"
  create_namespace = true

  set {
    name  = "installCRDs"
    value = "true"
  }

  depends_on = [module.external_secrets_identity]
}

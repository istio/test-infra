# Asymmetric signing key used by cosign to sign release artifacts. The release
# Prow job references it through COSIGN_KEY=awskms:///alias/istio-cosign (see
# prow/eks/config/jobs/.base.yaml).
#
# Access is scoped entirely through the key's resource policy: only the
# prowjob-release workload role may sign with, or read the public half of, this
# key. No other workload role is granted KMS permissions, so the effective
# signer set is the release job alone. Granting via the key policy (rather than
# the prowjob-release identity policy in iam.tf) keeps the dependency one-way:
# the key policy reads the release role ARN, avoiding a cycle.

data "aws_caller_identity" "current" {}

resource "aws_kms_key" "istio_cosign" {
  description              = "Asymmetric signing key for cosign release signatures."
  customer_master_key_spec = "ECC_NIST_P256"
  key_usage                = "SIGN_VERIFY"
  deletion_window_in_days  = 30

  policy = data.aws_iam_policy_document.istio_cosign_key.json
}

resource "aws_kms_alias" "istio_cosign" {
  name          = "alias/istio-cosign"
  target_key_id = aws_kms_key.istio_cosign.key_id
}

data "aws_iam_policy_document" "istio_cosign_key" {
  # Account root retains administrative control so the key stays manageable via
  # IAM; this does not grant signing to any workload (no role has kms:Sign).
  statement {
    sid       = "EnableIAMAdmin"
    effect    = "Allow"
    actions   = ["kms:*"]
    resources = ["*"]

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }

  # Only the release job may sign with / read this key.
  statement {
    sid    = "ReleaseSign"
    effect = "Allow"
    actions = [
      "kms:Sign",
      "kms:Verify",
      "kms:GetPublicKey",
      "kms:DescribeKey",
    ]
    resources = ["*"]

    principals {
      type        = "AWS"
      identifiers = [module.workload_identity["prowjob-release"].iam_role_arn]
    }
  }
}

# The EBS CSI controller needs block-volume management permissions only in the
# control-plane cluster, where ghproxy's persistent cache runs.
module "ebs_csi_identity" {
  source  = "terraform-aws-modules/eks-pod-identity/aws"
  version = "~> 2.0"

  name = "aws-ebs-csi-prow"

  attach_aws_ebs_csi_policy = true
}

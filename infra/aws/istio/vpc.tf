locals {
  azs = ["us-west-2a", "us-west-2b", "us-west-2c"]

  vpcs = {
    prow         = { cidr = "10.0.0.0/16" } # control plane / trusted
    prow-build   = { cidr = "10.1.0.0/16" } # core build (x86 + ARM)
    prow-private = { cidr = "10.2.0.0/16" } # private / PSWG build
  }
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 6.0"

  for_each = local.vpcs

  name = "${each.key}-vpc"
  cidr = each.value.cidr
  azs  = local.azs

  private_subnets = [for i in range(length(local.azs)) : cidrsubnet(each.value.cidr, 4, i)]
  public_subnets  = [for i in range(length(local.azs)) : cidrsubnet(each.value.cidr, 4, i + 8)]

  # a single NAT gateway keeps cost down. Revisit for HA later.
  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true

  # Subnet tags so EKS can discover where to place load balancers.
  private_subnet_tags = { "kubernetes.io/role/internal-elb" = "1" }
  public_subnet_tags  = { "kubernetes.io/role/elb" = "1" }
}

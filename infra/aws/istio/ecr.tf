resource "aws_ecr_pull_through_cache_rule" "ghcr" {
  ecr_repository_prefix = "ghcr"
  upstream_registry_url = "ghcr.io"
  credential_arn        = aws_secretsmanager_secret.secrets["ecr-pullthroughcache/ghcr"].arn
}

resource "aws_ecr_pull_through_cache_rule" "docker" {
  ecr_repository_prefix = "docker"
  upstream_registry_url = "registry-1.docker.io"
  credential_arn        = aws_secretsmanager_secret.secrets["ecr-pullthroughcache/docker-hub"].arn
}

resource "aws_ecr_pull_through_cache_rule" "quay" {
  ecr_repository_prefix = "quay"
  upstream_registry_url = "quay.io"
}

resource "aws_ecr_repository_creation_template" "ghcr" {
  applied_for = ["PULL_THROUGH_CACHE"]
  prefix      = "ghcr"

  lifecycle_policy = local.ecr_pull_through_cache_lifecycle_policy
}

resource "aws_ecr_repository_creation_template" "docker" {
  applied_for = ["PULL_THROUGH_CACHE"]
  prefix      = "docker"

  lifecycle_policy = local.ecr_pull_through_cache_lifecycle_policy
}

resource "aws_ecr_repository_creation_template" "quay" {
  applied_for = ["PULL_THROUGH_CACHE"]
  prefix      = "quay"

  lifecycle_policy = local.ecr_pull_through_cache_lifecycle_policy
}

locals {
  ecr_pull_through_cache_lifecycle_policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "Retain the 100 most recent cached images per repository"
      selection = {
        tagStatus   = "any"
        countType   = "imageCountMoreThan"
        countNumber = 100
      }
      action = {
        type = "expire"
      }
    }]
  })

  ecr_interface_endpoint_definitions = {
    for pair in setproduct(keys(local.vpcs), toset(["ecr.api", "ecr.dkr"])) :
    "${pair[0]}-${replace(pair[1], ".", "-")}" => {
      vpc     = pair[0]
      service = pair[1]
    }
  }
}

resource "aws_security_group" "ecr_endpoint_access_group" {
  for_each = local.vpcs

  name        = "${each.key}-ecr-interface-endpoints"
  description = "Security group attached to ${each.key} private ECR interface endpoints"
  vpc_id      = module.vpc[each.key].vpc_id
}

resource "aws_vpc_security_group_ingress_rule" "ecr_interface_endpoint_https_from_vpc" {
  for_each = local.vpcs

  security_group_id = aws_security_group.ecr_endpoint_access_group[each.key].id
  description       = "Allow HTTPS from ${each.key} workloads to private ECR interface endpoints"
  cidr_ipv4         = each.value.cidr
  from_port         = 443
  ip_protocol       = "tcp"
  to_port           = 443
}

resource "aws_vpc_endpoint" "ecr_interface" {
  for_each = local.ecr_interface_endpoint_definitions

  vpc_id              = module.vpc[each.value.vpc].vpc_id
  service_name        = "com.amazonaws.${local.region}.${each.value.service}"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = true
  subnet_ids          = module.vpc[each.value.vpc].private_subnets
  security_group_ids  = [aws_security_group.ecr_endpoint_access_group[each.value.vpc].id]
}

resource "aws_vpc_endpoint" "s3_gateway" {
  for_each = local.vpcs

  vpc_id            = module.vpc[each.key].vpc_id
  service_name      = "com.amazonaws.${local.region}.s3"
  vpc_endpoint_type = "Gateway"
  route_table_ids   = module.vpc[each.key].private_route_table_ids
}

terraform {
  required_version = ">= 1.10"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }

  # Using local state for now. Once the state bucket exists, add an S3 backend:
  #
  # backend "s3" {
  #   bucket       = "istio-terraform-aws"
  #   key          = "secrets/terraform.tfstate"
  #   region       = "us-west-2"
  #   encrypt      = true
  #   use_lockfile = true
  # }
}

provider "aws" {
  region = local.region

  default_tags {
    tags = {
      managed-by = "terraform"
      owner      = "istio"
    }
  }
}

terraform {
  required_version = ">= 1.10"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.17"
    }
  }

  backend "s3" {
    bucket       = "istio-terraform"
    key          = "istio/terraform.tfstate"
    region       = "us-west-2"
    encrypt      = true
    use_lockfile = true
  }
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

# Helm provider targets the control-plane (`prow`) cluster, where the ingress
# objects (deck, hook, gcsweb) live. Auth uses `aws eks get-token` so no
# long-lived credential is written to state.
provider "helm" {
  kubernetes {
    host                   = module.eks["prow"].cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks["prow"].cluster_certificate_authority_data)

    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args        = ["eks", "get-token", "--cluster-name", module.eks["prow"].cluster_name, "--region", local.region]
    }
  }
}

provider "helm" {
  alias = "prow_build"
  kubernetes {
    host                   = module.eks["prow-build"].cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks["prow-build"].cluster_certificate_authority_data)

    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args        = ["eks", "get-token", "--cluster-name", module.eks["prow-build"].cluster_name, "--region", local.region]
    }
  }
}

provider "helm" {
  alias = "prow_private"
  kubernetes {
    host                   = module.eks["prow-private"].cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks["prow-private"].cluster_certificate_authority_data)

    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args        = ["eks", "get-token", "--cluster-name", module.eks["prow-private"].cluster_name, "--region", local.region]
    }
  }
}

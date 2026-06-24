terraform {
  required_version = ">= 1.10"

  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.18.0"
    }
  }

  backend "s3" {
    bucket       = "istio-terraform"
    key          = "cf/cncf-istio/terraform.tfstate"
    region       = "us-west-2"
    encrypt      = true
    use_lockfile = true
  }
}

provider "cloudflare" {
}

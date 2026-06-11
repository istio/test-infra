terraform {
   backend "gcs" {
    bucket = "istio-terraform"
    prefix = "cf/cncf-istio" // project name
  }

  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.18.0"
    }
  }
}

provider "cloudflare" {
}

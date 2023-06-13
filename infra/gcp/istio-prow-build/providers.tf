terraform {
  backend "gcs" {
    bucket = "istio-terraform"
    prefix = "istio-prow-build" // project name
  }

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.69.1"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 4.69.1"
    }
  }
}

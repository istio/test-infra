terraform {
  backend "gcs" {
    bucket = "istio-terraform"
    prefix = "istio-testing" // project name
  }

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.20.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 5.20.0"
    }
  }
}

terraform {
  backend "gcs" {
    bucket = "istio-terraform"
    prefix = "istio-io" // project name
  }

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.9.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 6.9.0"
    }
  }
}

provider "google" {
  project = local.project_id
}

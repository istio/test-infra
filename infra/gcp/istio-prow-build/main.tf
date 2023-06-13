
locals {
  project_id = "istio-prow-build"
}

data "google_organization" "org" {
  domain = "google.com"
}

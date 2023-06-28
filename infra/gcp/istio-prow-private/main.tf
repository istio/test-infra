locals {
  project_id     = "istio-prow-private"
  project_number = "725673368965"

  pod_namespace = "test-pods"
}
provider "google" {
  project = local.project_id
}

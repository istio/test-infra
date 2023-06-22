resource "google_artifact_registry_repository" "main" {
  location      = "us"
  repository_id = "istio-prow-private"
  description   = "registry to host private Istio release artifacts"
  format        = "DOCKER"

  docker_config {
    immutable_tags = true
  }
}

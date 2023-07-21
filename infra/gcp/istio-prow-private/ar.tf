resource "google_artifact_registry_repository" "main" {
  location      = "us"
  repository_id = "istio-prow-private"
  description   = "registry to host private Istio release artifacts"
  format        = "DOCKER"

  docker_config {
    immutable_tags = false
  }

  lifecycle {
    prevent_destroy = true
  }
}

data "google_iam_policy" "artifact_registry" {
  binding {
    role = "roles/artifactregistry.createOnPushWriter"
    members = [
      "serviceAccount:${module.prowjob_private_account.email}",
    ]
  }
  binding {
    role = "roles/artifactregistry.reader"
    members = [
      // Give testers access to artifacts. TODO: move to google groups
      "user:sshapar@google.com",
    ]
  }
}
resource "google_artifact_registry_repository_iam_policy" "policy" {
  project     = google_artifact_registry_repository.main.project
  location    = google_artifact_registry_repository.main.location
  repository  = google_artifact_registry_repository.main.name
  policy_data = data.google_iam_policy.artifact_registry.policy_data
}

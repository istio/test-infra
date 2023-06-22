# secrets.tf contains GCP managed secrets
# This file just defines the secrets and their access; the actual secrets are managed manually:
# gcloud --project <project> secrets versions add <secret> --data-file=<data>

# This is just a test secret for testing
resource "google_secret_manager_secret" "test_secret" {
  project   = local.project_id
  secret_id = "test-secret"
  replication {
    automatic = true
  }
}
resource "google_secret_manager_secret" "test_secret2" {
  project   = local.project_id
  secret_id = "test-secret2"
  replication {
    automatic = true
  }
}

data "google_iam_policy" "test_secret" {
  binding {
    role = "roles/secretmanager.secretAccessor"
    members = [
      "serviceAccount:${module.prowjob_release_account.email}",
    ]
  }
}

resource "google_secret_manager_secret_iam_policy" "test_secret" {
  project     = google_secret_manager_secret.test_secret.project
  secret_id   = google_secret_manager_secret.test_secret.secret_id
  policy_data = data.google_iam_policy.test_secret.policy_data
}
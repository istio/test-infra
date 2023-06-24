# secrets.tf contains GCP managed secrets

locals {
  # release_secrets contains secrets for prowjob-release
  release_secrets = [
    # Access token for "istio" dockerhub account
    "release_docker_istio",
    # Access token for Github "istio-release-robot" with scopes: admin:repo_hook, notifications, repo, workflow
    "release_github_istio-release",
    # Access token for Grafana for the "Istio" org. Named "release-pipeline-token" in Grafana, with role "Editor".
    "release_grafana_istio",
  ]
  # github_read_secret contains secrets for prowjob-github-read
  github_readonly_secret = [
    # Fine grained PAT in the Istio org, "github/release-notes/public-read-only". Has public access only.
    "github-read_github_read" # Expires on 6/23/2024. TODO: find the best way to ensure this is noted.
  ]

  all_secrets = concat(
    local.release_secrets,
    local.github_readonly_secret
  )
}

# Create all secrets. This just creates the secret resource; the actual secrets are managed manually:
# gcloud --project <project> secrets versions add <secret> --data-file=<data>
resource "google_secret_manager_secret" "secrets" {
  for_each  = toset(local.all_secrets)
  project   = local.project_id
  secret_id = each.key
  replication {
    automatic = true
  }
}

# For each "release secret", give access to the release job.
data "google_iam_policy" "release_secret_access" {
  binding {
    role = "roles/secretmanager.secretAccessor"
    members = [
      "serviceAccount:${module.prowjob_release_account.email}",
    ]
  }
}
resource "google_secret_manager_secret_iam_policy" "release_secret_access" {
  for_each    = toset(local.release_secrets)
  project     = local.project_id
  secret_id   = each.key
  policy_data = data.google_iam_policy.release_secret_access.policy_data
}

# For each "github read secret", give access to the release job.
data "google_iam_policy" "github_readonly_access" {
  binding {
    role = "roles/secretmanager.secretAccessor"
    members = [
      "serviceAccount:${module.prowjob_github_read_account.email}",
    ]
  }
}
resource "google_secret_manager_secret_iam_policy" "github_readonly_access" {
  for_each    = toset(local.github_readonly_secret)
  project     = local.project_id
  secret_id   = each.key
  policy_data = data.google_iam_policy.github_readonly_access.policy_data
  depends_on  = [google_secret_manager_secret.secrets]
}

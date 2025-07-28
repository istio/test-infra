# secrets.tf contains GCP managed secrets

locals {
  # release_secrets contains secrets for prowjob-release
  all_secrets = [
    # Access token for "istio" dockerhub account
    "release_docker_istio",
    # Fine grained PAT in the Istio org, "github/istio-release/release". Has write access to "Contents" and "Workflows".
    # Expires 7/29/2026.
    "release_github_istio-release",
    # Access token for Grafana for the "Istio" org. Named "release-pipeline-token" in Grafana, with role "Editor".
    "release_grafana_istio",

    # Fine grained PAT in the Istio org, "github/release-notes/public-read-only". Has public access only.
    "github-read_github_read", # Expires on 5/15/2025. TODO: find the best way to ensure this is noted.

    # Classic PAT for user "istio-testing", token name "github/istio-testing/pusher". Has scopes `repo,read:user`.
    # TODO(https://github.com/orgs/community/discussions/36441#discussioncomment-6286043) Use fine grained tokens
    "github_istio-testing_pusher", # No expiration
  ]
}

# Create all secrets. This just creates the secret resource; the actual secrets are managed manually:
# gcloud --project <project> secrets versions add <secret> --data-file=<data>
resource "google_secret_manager_secret" "secrets" {
  for_each  = toset(local.all_secrets)
  project   = local.project_id
  secret_id = each.key
  replication {
    auto {}
  }
}

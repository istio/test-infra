locals {
  # release_secrets contains secrets for the release prowjob.
  all_secrets = {
    # Access token for "istio" dockerhub account.
    "release_docker_istio" = "DockerHub access token for the \"istio\" account"

    # Fine grained PAT in the Istio org, "github/istio-release/release".
    # Has write access to "Contents" and "Workflows". Expires 7/29/2026.
    "release_github_istio-release" = "Fine-grained GitHub PAT for releases (Contents+Workflows write); expires 2026-07-29"

    # Access token for Grafana for the "Istio" org. Named
    # "release-pipeline-token" in Grafana, with role "Editor".
    "release_grafana_istio" = "Grafana token for the Istio org (Editor role)"

    # Fine grained PAT in the Istio org, "github/release-notes/public-read-only".
    # Has public access only. Expires on 5/19/2027.
    "github-read_github_read" = "Fine-grained GitHub PAT, public read-only; expires 2027-05-19"

    # Classic PAT for user "istio-testing", token name
    # "github/istio-testing/pusher". Has scopes `repo,read:user`. No expiration.
    "github_istio-testing_pusher" = "Classic GitHub PAT for istio-testing (repo,read:user); no expiration"

    # Permanent Cloudflare admin token; used to mint the ephemeral R2 creds.
    "cf_r2_admin_token" = "Permanent Cloudflare admin token (mints ephemeral R2 credentials)"

    # Generic read-only R2 credentials for the public buckets.
    "cf_r2_public_buckets_ro_credentials" = "Read-only R2 credentials for public buckets"

    # GitHub OAuth token for Prow jobs (JSON: {oauth}).
    "oauth_token" = "GitHub OAuth token for Prow jobs"

    # HMAC token used by hook to validate GitHub webhook payloads.
    "hmac_token" = "HMAC token for validating GitHub webhooks"

    # Cookie secret used by deck/deck-private to sign session cookies.
    "cookie" = "Cookie secret for signing Prow deck session cookies"

    # GitHub OAuth app config for the public deck login flow.
    "github-oauth-config" = "GitHub OAuth app config for deck login"

    # GitHub OAuth app config for the private deck login flow.
    "github-oauth-config-private" = "GitHub OAuth app config for deck-private login"

    # Slack token used by crier/hook for Slack notifications.
    "slack_token" = "Slack token for Prow notifications"

    # OAuth2-proxy client credentials guarding deck-private.
    "deck-oauth-proxy" = "OAuth2-proxy client credentials for deck-private"

    # SSH key material for the istio-testing robot (JSON: {id_rsa, id_rsa.pub, known_hosts}).
    "istio-testing_robot-ssh-key" = "SSH key for the istio-testing robot"

    # Bucket specific R2 credentials
    "cf_r2_istio-release_credentials"            = "Ephemeral Cloudflare R2 credentials for the istio-release bucket"
    "cf_r2_istio-build_credentials"              = "Ephemeral Cloudflare R2 credentials for the istio-build bucket"
    "cf_r2_istio-build-private_credentials"      = "Ephemeral Cloudflare R2 credentials for the istio-build-private bucket"
    "cf_r2_istio-prerelease_credentials"         = "Ephemeral Cloudflare R2 credentials for the istio-prerelease bucket"
    "cf_r2_istio-prerelease-private_credentials" = "Ephemeral Cloudflare R2 credentials for the istio-prerelease-private bucket"
    "cf_r2_istio-prow_credentials"               = "Ephemeral Cloudflare R2 credentials for the istio-prow bucket"
    "cf_r2_istio-prow-private_credentials"       = "Ephemeral Cloudflare R2 credentials for the istio-prow-private bucket"
    "cf_r2_istio-testgrid_credentials"           = "Ephemeral Cloudflare R2 credentials for the istio-testgrid bucket"

    # Read-only S3 credentials for gcsweb (public buckets). The value is managed
    # by Terraform (aws_secretsmanager_secret_version.gcsweb_s3_credentials).
    "gcsweb_s3_credentials" = "Read-only S3 credentials for gcsweb (public buckets)"

    "ghcr-push-pat" = "PAT to push to ghcr.io"
  }
}

resource "aws_secretsmanager_secret" "secrets" {
  for_each    = local.all_secrets
  name        = each.key
  description = each.value
}

# Read-only S3 credentials for gcsweb, in the format expected by the prow
# storage provider's s3-credentials file. The access key is minted in iam.tf
# and synced into the cluster via external_secrets.yaml.
resource "aws_secretsmanager_secret_version" "gcsweb_s3_credentials" {
  secret_id = aws_secretsmanager_secret.secrets["gcsweb_s3_credentials"].id
  secret_string = jsonencode({
    region              = "us-west-2"
    s3_force_path_style = false
    access_key          = aws_iam_access_key.gcsweb.id
    secret_key          = aws_iam_access_key.gcsweb.secret
  })
}

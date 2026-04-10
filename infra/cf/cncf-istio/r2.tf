resource "cloudflare_r2_bucket" "istio-release" {
  account_id = var.account_id
  name       = "istio-release"
}

resource "cloudflare_r2_custom_domain" "istio-release" {
  account_id  = var.account_id
  domain      = "istio-release.r2.istio.io"
  enabled     = true
  bucket_name = cloudflare_r2_bucket.istio-release.name
  zone_id     = var.zone_id
}

resource "cloudflare_r2_bucket" "istio-testgrid" {
  account_id = var.account_id
  name       = "istio-testgrid"
}

resource "cloudflare_r2_custom_domain" "istio-testgrid" {
  account_id  = var.account_id
  domain      = "istio-testgrid.r2.istio.io"
  enabled     = true
  bucket_name = cloudflare_r2_bucket.istio-testgrid.name
  zone_id     = var.zone_id
}

resource "cloudflare_r2_bucket" "istio-build" {
  account_id = var.account_id
  name       = "istio-build"
}

resource "cloudflare_r2_custom_domain" "istio-build" {
  account_id  = var.account_id
  domain      = "istio-build.r2.istio.io"
  enabled     = true
  bucket_name = cloudflare_r2_bucket.istio-build.name
  zone_id     = var.zone_id
}

resource "cloudflare_r2_bucket" "istio-build-private" {
  account_id = var.account_id
  name       = "istio-build-private"
}

resource "cloudflare_r2_bucket" "istio-prow" {
  account_id = var.account_id
  name       = "istio-prow"
}

resource "cloudflare_r2_custom_domain" "istio-prow" {
  account_id  = var.account_id
  domain      = "istio-prow.r2.istio.io"
  enabled     = true
  bucket_name = cloudflare_r2_bucket.istio-prow.name
  zone_id     = var.zone_id
}

resource "cloudflare_r2_bucket" "istio-prow-private" {
  account_id = var.account_id
  name       = "istio-prow-private"
}

resource "cloudflare_r2_bucket" "istio-prerelease" {
  account_id = var.account_id
  name       = "istio-prerelease"
}

resource "cloudflare_r2_custom_domain" "istio-prerelease" {
  account_id  = var.account_id
  domain      = "istio-prerelease.r2.istio.io"
  enabled     = true
  bucket_name = cloudflare_r2_bucket.istio-prerelease.name
  zone_id     = var.zone_id
}

resource "cloudflare_r2_bucket" "istio-prerelease-private" {
  account_id = var.account_id
  name       = "istio-prerelease-private"
}

locals {
  r2_public_buckets = toset([
    cloudflare_r2_bucket.istio-build.name,
    cloudflare_r2_bucket.istio-prow.name,
    cloudflare_r2_bucket.istio-prerelease.name,
    cloudflare_r2_bucket.istio-release.name,
    cloudflare_r2_bucket.istio-testgrid.name,
  ])
}

resource "cloudflare_ruleset" "blob_redirect" {
  zone_id     = var.zone_id
  name        = "blob-redirect"
  description = "Redirect blob.istio.io/<bucket>/path to <bucket>.r2.istio.io/path"
  kind        = "zone"
  phase       = "http_request_dynamic_redirect"

  rules = [for bucket in local.r2_public_buckets : {
    ref         = "blob_redirect_${bucket}"
    description = "Redirect blob.istio.io/${bucket}/* to ${bucket}.r2.istio.io/*"
    expression  = "(http.host eq \"blob.istio.io\" and starts_with(http.request.uri.path, \"/${bucket}/\"))"
    action      = "redirect"
    action_parameters = {
      from_value = {
        status_code = 301
        target_url = {
          expression = "concat(\"https://${bucket}.r2.istio.io\", regex_replace(http.request.uri.path, \"^/${bucket}/\", \"/\"))"
        }
        preserve_query_string = true
      }
    }
  }]
}

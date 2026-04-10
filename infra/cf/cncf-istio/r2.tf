resource "cloudflare_r2_bucket" "istio-release" {
  account_id = var.account_id
  name       = "istio-release"
}

resource "cloudflare_r2_custom_domain" "istio-release" {
  account_id = var.account_id
  hostname = "istio-release.r2.istio.io"
  enabled = true
  bucket_name = cloudflare_r2_bucket.istio-release.name
  zone_id = var.zone_id
}

resource "cloudflare_r2_bucket" "istio-testgrid" {
  account_id = var.account_id
  name       = "istio-testgrid"
}

resouce "cloudflare_r2_custom_domain" "istio-testgrid" {
  account_id = var.account_id
  hostname = "istio-testgrid.r2.istio.io"
  enabled = true
  bucket_name = cloudflare_r2_bucket.istio-testgrid.name
  zone_id = var.zone_id
}

resource "cloudflare_r2_bucket" "istio-build" {
  account_id = var.account_id
  name       = "istio-build"
}

resource "cloudflare_r2_custom_domain" "istio-build" {
  account_id = var.account_id
  hostname = "istio-build.r2.istio.io"
  enabled = true
  bucket_name = cloudflare_r2_bucket.istio-build.name
  zone_id = var.zone_id
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
  account_id = var.account_id
  hostname = "istio-prow.r2.istio.io"
  enabled = true
  bucket_name = cloudflare_r2_bucket.istio-prow.name
  zone_id = var.zone_id
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
  account_id = var.account_id
  hostname = "istio-prerelease.r2.istio.io"
  enabled = true
  bucket_name = cloudflare_r2_bucket.istio-prerelease.name
  zone_id = var.zone_id
}

resource "cloudflare_r2_bucket" "istio-prerelease-private" {
  account_id = var.account_id
  name       = "istio-prerelease-private"
}

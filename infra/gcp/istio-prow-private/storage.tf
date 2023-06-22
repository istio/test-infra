# Mirrors the gs://istio-build bucket
resource "google_storage_bucket" "istio_build_private" {
  name          = "istio-build-private"
  location      = "US"
  storage_class = "STANDARD"

  uniform_bucket_level_access = true
  lifecycle {
    prevent_destroy = true
  }
}
# Mirrors the gs://istio-prerelease bucket
resource "google_storage_bucket" "istio_prerelease_private" {
  name          = "istio-prerelease-private"
  location      = "US"
  storage_class = "STANDARD"

  uniform_bucket_level_access = true
  lifecycle {
    prevent_destroy = true
  }
}

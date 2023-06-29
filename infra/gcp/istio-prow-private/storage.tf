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
# Mirrors the gs://istio-prow bucket, which stores Prow logs
resource "google_storage_bucket" "istio_prow_private" {
  name          = "istio-prow-private"
  location      = "US"
  storage_class = "STANDARD"

  uniform_bucket_level_access = true
  lifecycle {
    prevent_destroy = true
  }
}
# Give control plane (deck-private) read access to the bucket so logs can be read
resource "google_storage_bucket_iam_member" "istio_prow_private_deck" {
  bucket = google_storage_bucket.istio_prow_private.name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:prow-control-plane@istio-testing.iam.gserviceaccount.com"
}

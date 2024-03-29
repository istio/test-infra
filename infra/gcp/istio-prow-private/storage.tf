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
# Give testers access to artifacts. TODO: move to google groups
resource "google_storage_bucket_iam_member" "istio_prerelease_private" {
  bucket = google_storage_bucket.istio_prerelease_private.name
  role   = "roles/storage.objectViewer"
  member = "user:sshapar@google.com"
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
# Give control plane access to the bucket
# Deck: logs can be read
# Crier: writes various artifacts like finished.json, etc
resource "google_storage_bucket_iam_member" "istio_prow_private_deck" {
  bucket = google_storage_bucket.istio_prow_private.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:prow-control-plane@istio-testing.iam.gserviceaccount.com"
}

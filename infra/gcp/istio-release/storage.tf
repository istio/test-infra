# PRODUCTION CRITICAL: hosts gcr.io/istio-release
resource "google_storage_bucket" "artifacts_istio_release_appspot_com" {
  force_destroy = false
  location      = "US"
  name          = "artifacts.istio-release.appspot.com"
  project       = "istio-release"
  storage_class = "STANDARD"
}

# This is a legacy bucket from our release process before istio/release-builder.
# It could *probably* be cleaned up, but safer to keep it around.
resource "google_storage_bucket" "istio_release_pipeline_data" {
  force_destroy            = false
  location                 = "US"
  name                     = "istio-release-pipeline-data"
  project                  = "istio-release"
  public_access_prevention = "inherited"
  storage_class            = "MULTI_REGIONAL"
}

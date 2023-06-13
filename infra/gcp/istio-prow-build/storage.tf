resource "google_storage_bucket" "artifacts_istio_prow_build_appspot_com" {
  force_destroy               = false
  location                    = "US"
  name                        = "artifacts.istio-prow-build.appspot.com"
  project                     = "istio-prow-build"
  public_access_prevention    = "inherited"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "istio_private_artifacts" {
  force_destroy               = false
  location                    = "US"
  name                        = "istio-private-artifacts"
  project                     = "istio-prow-build"
  public_access_prevention    = "inherited"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "istio_private_build" {
  force_destroy               = false
  location                    = "US"
  name                        = "istio-private-build"
  project                     = "istio-prow-build"
  public_access_prevention    = "inherited"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "istio_private_prerelease" {
  force_destroy               = false
  location                    = "US"
  name                        = "istio-private-prerelease"
  project                     = "istio-prow-build"
  public_access_prevention    = "inherited"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}

# CRITICAL: hosts gcr.io/istio-prerelease-testing
resource "google_storage_bucket" "artifacts_istio_prerelease_testing_appspot_com" {
  force_destroy            = false
  location                 = "US"
  name                     = "artifacts.istio-prerelease-testing.appspot.com"
  project                  = "istio-prerelease-testing"
  public_access_prevention = "inherited"
  storage_class            = "STANDARD"
}

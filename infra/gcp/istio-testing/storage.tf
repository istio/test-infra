resource "google_storage_bucket" "artifacts_istio_testing_appspot_com" {
  force_destroy = false
  location      = "US"
  name          = "artifacts.istio-testing.appspot.com"
  project       = "istio-testing"
  storage_class = "STANDARD"
}
resource "google_storage_bucket" "e2e_testing_log" {
  force_destroy = false
  location      = "US"
  name          = "e2e-testing-log"
  project       = "istio-testing"
  storage_class = "MULTI_REGIONAL"
}
resource "google_storage_bucket" "istio_artifacts" {
  force_destroy = false
  location      = "US"
  name          = "istio-artifacts"
  project       = "istio-testing"
  storage_class = "MULTI_REGIONAL"
}
resource "google_storage_bucket" "istio_build_deps" {
  force_destroy = false
  location      = "US"
  name          = "istio-build-deps"
  project       = "istio-testing"
  storage_class = "MULTI_REGIONAL"
}
resource "google_storage_bucket" "istio_build" {
  force_destroy = false
  location      = "US"
  name          = "istio-build"
  project       = "istio-testing"
  storage_class = "MULTI_REGIONAL"
}
resource "google_storage_bucket" "istio_circleci" {
  force_destroy            = false
  location                 = "US"
  name                     = "istio-circleci"
  project                  = "istio-testing"
  public_access_prevention = "inherited"
  storage_class            = "MULTI_REGIONAL"

  versioning {
    enabled = true
  }
}
resource "google_storage_bucket" "istio_code_coverage" {
  force_destroy = false
  location      = "US"
  name          = "istio-code-coverage"
  project       = "istio-testing"
  storage_class = "MULTI_REGIONAL"
}
resource "google_storage_bucket" "istio_flakes" {
  force_destroy            = false
  location                 = "US"
  name                     = "istio-flakes"
  project                  = "istio-testing"
  public_access_prevention = "inherited"
  storage_class            = "MULTI_REGIONAL"
}
resource "google_storage_bucket" "istio_gists" {
  force_destroy = false
  location      = "US"
  name          = "istio-gists"
  project       = "istio-testing"
  storage_class = "MULTI_REGIONAL"
}
resource "google_storage_bucket" "istio_logs" {
  force_destroy = false
  location      = "US"
  name          = "istio-logs"
  project       = "istio-testing"
  storage_class = "MULTI_REGIONAL"
}
resource "google_storage_bucket" "istio_presubmit_release_pipeline_data" {
  force_destroy               = false
  location                    = "US"
  name                        = "istio-presubmit-release-pipeline-data"
  project                     = "istio-testing"
  public_access_prevention    = "inherited"
  storage_class               = "MULTI_REGIONAL"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "istio_prow" {
  force_destroy = false
  location      = "US"
  name          = "istio-prow"
  project       = "istio-testing"
  storage_class = "MULTI_REGIONAL"
}
resource "google_storage_bucket" "istio_stats" {
  force_destroy            = false
  location                 = "US"
  name                     = "istio-stats"
  project                  = "istio-testing"
  public_access_prevention = "inherited"
  storage_class            = "MULTI_REGIONAL"
}
resource "google_storage_bucket" "istio_testgrid" {
  force_destroy               = false
  location                    = "US"
  name                        = "istio-testgrid"
  project                     = "istio-testing"
  public_access_prevention    = "enforced"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "istio_testing_cloudbuild" {
  force_destroy = false
  location      = "US"
  name          = "istio-testing_cloudbuild"
  project       = "istio-testing"
  storage_class = "STANDARD"
}
resource "google_storage_bucket" "istio_tools" {
  force_destroy = false
  location      = "US"
  name          = "istio-tools"
  project       = "istio-testing"
  storage_class = "MULTI_REGIONAL"
}
resource "google_storage_bucket" "us_artifacts_istio_testing_appspot_com" {
  force_destroy = false
  location      = "US"
  name          = "us.artifacts.istio-testing.appspot.com"
  project       = "istio-testing"
  storage_class = "STANDARD"
}
resource "google_storage_bucket" "istio_snippets" {
  force_destroy            = false
  location                 = "US-WEST1"
  name                     = "istio-snippets"
  project                  = "istio-testing"
  public_access_prevention = "inherited"
  storage_class            = "STANDARD"
}

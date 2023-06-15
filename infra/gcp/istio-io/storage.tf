# CRITICAL: stores prerelease build information.
# istio/release-builder will push builds of releases here. Once validated they are pushed to "istio-release".
resource "google_storage_bucket" "istio_prerelease" {
  force_destroy = false
  location      = "US"
  name          = "istio-prerelease"
  project       = "istio-io"
  storage_class = "MULTI_REGIONAL"
}

# PRODUCTION CRITICAL: stores our official releases.
resource "google_storage_bucket" "istio_release" {
  force_destroy = false
  location      = "US"
  name          = "istio-release"
  project       = "istio-io"
  storage_class = "MULTI_REGIONAL"

  versioning {
    enabled = true
  }
}

# Not managed by terraform: "istio-terraform"
# This is where our terraform state is stored.
# I think you can technically manage this with terraform, but it feels a bit circular so it is avoided for now
# terraform import google_storage_bucket.istio_terraform istio-terraform

# Stores data for fortio runs. Published at fortio.istio.io.
# This hasn't actually been used for many years, but it does still have publicly accessible data.
resource "google_storage_bucket" "fortio_data" {
  force_destroy            = false
  location                 = "US-WEST1"
  name                     = "fortio-data"
  project                  = "istio-io"
  public_access_prevention = "inherited"
  storage_class            = "REGIONAL"
}

# gcr.io/istio-io. Schedule for removal in the near future, currently it hosts images from ~2017.
resource "google_storage_bucket" "artifacts_istio_io_appspot_com" {
  force_destroy = false
  location      = "US"
  name          = "artifacts.istio-io.appspot.com"
  project       = "istio-io"
  storage_class = "STANDARD"
}

# Legacy bucket from when cloudbuild was used ~2017. Scheduled for removal in the future.
resource "google_storage_bucket" "istio_io_cloudbuild" {
  force_destroy = false
  location      = "US"
  name          = "istio-io_cloudbuild"
  project       = "istio-io"
  storage_class = "STANDARD"
}

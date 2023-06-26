resource "google_storage_bucket" "artifacts_istio_prow_build_appspot_com" {
  force_destroy               = false
  location                    = "US"
  name                        = "artifacts.istio-prow-build.appspot.com"
  project                     = "istio-prow-build"
  public_access_prevention    = "inherited"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
data "google_iam_policy" "artifacts_istio_prow_build_appspot_com" {
  binding {
    role = "roles/storage.admin"
    members = [
      "group:istio-testing-admins@twosync.google.com",
      "serviceAccount:prow-internal-storage@istio-prow-build.iam.gserviceaccount.com",
      "serviceAccount:istio-prow-test-job-private@istio-testing.iam.gserviceaccount.com",
    ]
  }
  binding {
    members = [
      "projectOwner:istio-prow-build",
    ]
    role = "roles/storage.legacyBucketOwner"
  }
}
resource "google_storage_bucket_iam_policy" "artifacts_istio_prow_build_appspot_com" {
  bucket      = google_storage_bucket.artifacts_istio_prow_build_appspot_com.name
  policy_data = data.google_iam_policy.artifacts_istio_prow_build_appspot_com.policy_data
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
data "google_iam_policy" "istio_private_artifacts" {
  binding {
    role = "roles/storage.admin"
    members = [
      "group:istio-testing-admins@twosync.google.com",
      "serviceAccount:prow-internal-storage@istio-prow-build.iam.gserviceaccount.com",
    ]
  }
  binding {
    members = [
      "projectOwner:istio-prow-build",
    ]
    role = "roles/storage.legacyBucketOwner"
  }
  binding {
    members = [
      "serviceAccount:istio-prow-test-job-private@istio-testing.iam.gserviceaccount.com",
    ]
    role = "roles/storage.objectAdmin"
  }
}
resource "google_storage_bucket_iam_policy" "istio_private_artifacts" {
  bucket      = google_storage_bucket.istio_private_artifacts.name
  policy_data = data.google_iam_policy.istio_private_artifacts.policy_data
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
data "google_iam_policy" "istio_private_build" {
  binding {
    role = "roles/storage.admin"
    members = [
      "group:istio-testing-admins@twosync.google.com",
      "serviceAccount:prow-internal-storage@istio-prow-build.iam.gserviceaccount.com",
    ]
  }
  binding {
    members = [
      "projectOwner:istio-prow-build",
    ]
    role = "roles/storage.legacyBucketOwner"
  }
  binding {
    members = [
      "projectOwner:istio-prow-build",
    ]
    role = "roles/storage.legacyObjectOwner"
  }
  binding {
    members = [
      "serviceAccount:istio-prow-test-job-private@istio-testing.iam.gserviceaccount.com",
      "serviceAccount:prow-control-plane@istio-testing.iam.gserviceaccount.com",
    ]
    role = "roles/storage.objectAdmin"
  }
}
resource "google_storage_bucket_iam_policy" "istio_private_build" {
  bucket      = google_storage_bucket.istio_private_build.name
  policy_data = data.google_iam_policy.istio_private_build.policy_data
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
data "google_iam_policy" "istio_private_prerelease" {
  binding {
    role = "roles/storage.admin"
    members = [
      "group:istio-testing-admins@twosync.google.com",
      "serviceAccount:prow-internal-storage@istio-prow-build.iam.gserviceaccount.com",
    ]
  }
  binding {
    members = [
      "projectOwner:istio-prow-build",
    ]
    role = "roles/storage.legacyBucketOwner"
  }
  binding {
    members = [
      "projectOwner:istio-prow-build",
    ]
    role = "roles/storage.legacyObjectOwner"
  }
  binding {
    members = [
      "serviceAccount:istio-prow-test-job-private@istio-testing.iam.gserviceaccount.com",
    ]
    role = "roles/storage.objectAdmin"
  }
}
resource "google_storage_bucket_iam_policy" "istio_private_prerelease" {
  bucket      = google_storage_bucket.istio_private_prerelease.name
  policy_data = data.google_iam_policy.istio_private_prerelease.policy_data
}

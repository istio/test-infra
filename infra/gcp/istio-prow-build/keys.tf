resource "google_kms_key_ring" "istio_cosign_keyring" {
  location = "global"
  name     = "istio-cosign-keyring"
  project  = "istio-prow-build"
}

resource "google_kms_crypto_key" "istio_cosign_key" {
  destroy_scheduled_duration = "86400s"
  key_ring                   = "projects/istio-prow-build/locations/global/keyRings/istio-cosign-keyring"
  name                       = "istio-cosign-key"
  purpose                    = "ASYMMETRIC_SIGN"

  version_template {
    algorithm        = "EC_SIGN_P256_SHA256"
    protection_level = "SOFTWARE"
  }
}

# Only release job can use the keyring
data "google_iam_policy" "admin" {
  binding {
    role    = "roles/cloudkms.signerVerifier"
    members = ["serviceAccount:${module.prowjob_release_account.email}", ]
  }
  binding {
    role    = "roles/cloudkms.viewer"
    members = ["serviceAccount:${module.prowjob_release_account.email}", ]
  }
}
resource "google_kms_key_ring_iam_policy" "key_ring" {
  key_ring_id = google_kms_key_ring.istio_cosign_keyring.id
  policy_data = data.google_iam_policy.admin.policy_data
}
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

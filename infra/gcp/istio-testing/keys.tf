# TODO: the same-named keyring exists in istio-prow-build. Likely only one is used and we should drop the other.
# I suspect this is the unused one.
resource "google_kms_key_ring" "istio_cosign_keyring" {
  location = "global"
  name     = "istio-cosign-keyring"
  project  = "istio-testing"
}
resource "google_kms_crypto_key" "istio_cosign_key" {
  destroy_scheduled_duration = "86400s"
  key_ring                   = "projects/istio-testing/locations/global/keyRings/istio-cosign-keyring"
  name                       = "istio-cosign-key"
  purpose                    = "ASYMMETRIC_SIGN"

  version_template {
    algorithm        = "EC_SIGN_P256_SHA256"
    protection_level = "SOFTWARE"
  }
}

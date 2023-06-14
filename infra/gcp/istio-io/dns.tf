# PRODUCTION CRITICAL: the main zone for istio.io
# This is just the zone, we should find out how to get all the record sets defined as well
resource "google_dns_managed_zone" "istio_io" {
  dns_name      = "istio.io."
  description   = "Primary zone for istio.io"
  force_destroy = false
  name          = "istio-io"
  project       = "istio-io"
  visibility    = "public"

  # Not sure why we need all of this when its "off" anyways, but oh well..
  dnssec_config {
    kind          = "dns#managedZoneDnsSecConfig"
    non_existence = "nsec3"
    state         = "off"

    default_key_specs {
      algorithm  = "rsasha256"
      key_length = 2048
      key_type   = "keySigning"
      kind       = "dns#dnsKeySpec"
    }
    default_key_specs {
      algorithm  = "rsasha256"
      key_length = 1024
      key_type   = "zoneSigning"
      kind       = "dns#dnsKeySpec"
    }
  }
}

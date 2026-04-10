resource "cloudflare_dns_record" "site_verification_txt" {
  for_each = toset([
    "\"google-site-verification=tIhgL7XTBmvCOqWhMtTyp6oyf3YvZ6b3H7TFtcBNEYM\"",
    "\"google-site-verification=aQiMpyeG_m4y3cBlMJDpSvU7Mq1JwX6FWXJWiJrWFUM\"",
    "\"v=spf1 include:_spf.google.com ~all\"",
    "\"google-site-verification=RGVq8IS06qw1QOD0SXqQP12X9Rjv0tVd0xHjbHGguH4\"",
    "\"google-site-verification=-ICpI7h_HJD4LK3eEXurYzVrXOXBtQ6-cdsqi_SghBA\"",
    "\"google-gws-recovery-domain-verification=44243318\"",
  ])

  zone_id = var.zone_id
  name    = "@"
  ttl     = 3600
  type    = "TXT"
  content = each.key
}

resource "cloudflare_dns_record" "atproto_txt" {
  zone_id = var.zone_id
  name    = "_atproto"
  ttl     = 300
  type    = "TXT"
  content = "\"did=did:plc:gx2kyhr4aahbqzeyyjouvdo5\""
}

resource "cloudflare_dns_record" "spf_deny_txt" {
  for_each = toset([
    "gcsweb",
    "prow",
  ])

  zone_id = var.zone_id
  name    = each.key
  ttl     = 300
  type    = "TXT"
  content = "\"v=spf1 -all\""
}

resource "cloudflare_dns_record" "acme_challenge_txt" {
  for_each = toset([
    "\"gI0G5L6_SFkbQABaD2kKY3A8gS3EHFatwFxEWOHWiCc\"",
    "\"EAzsDPXrDUIBiXUcCD3qRzIvu1-bUW9xxl-b8QvsM90\""
  ])

  zone_id = var.zone_id
  name    = "_acme-challenge.eng"
  ttl     = 120
  type    = "TXT"
  content = each.key
}

resource "cloudflare_dns_record" "istio_mx" {
  for_each = {
    "aspmx.l.google.com."      = 1
    "alt1.aspmx.l.google.com." = 5
    "alt2.aspmx.l.google.com." = 5
    "alt3.aspmx.l.google.com." = 10
    "alt4.aspmx.l.google.com." = 10
  }

  zone_id  = var.zone_id
  name     = "@"
  ttl      = 3600
  type     = "MX"
  priority = each.value
  content  = each.key
}

resource "cloudflare_dns_record" "cname_records" {
  for_each = {
    "archive"       = "archive-istio.netlify.com."
    "elections"     = "router-default.apps.ospo-osci.z3b1.p1.openshiftapps.com."
    "em535"         = "u10461948.wl199.sendgrid.net."
    "em9813"        = "u17991541.wl042.sendgrid.net."
    "fortio"        = "istio.fortio.org."
    "latest"        = "istio-latest.netlify.app."
    "preliminary"   = "preliminary-istio.netlify.com."
    "s1._domainkey" = "s1.domainkey.u17991541.wl042.sendgrid.net."
    "s2._domainkey" = "s2.domainkey.u17991541.wl042.sendgrid.net."
    "slack"         = "ghs.googlehosted.com."
    "www"           = "istio.io."
  }

  zone_id = var.zone_id
  name    = each.key
  ttl     = 300
  type    = "CNAME"
  content = each.value
}

// Dummy AAAA record to enable Cloudflare proxy for blob.istio.io, which is required to trigger redirect rules to <bucket>.r2.istio.io.
resource "cloudflare_dns_record" "blob" {
  zone_id = var.zone_id
  name    = "blob"
  type    = "AAAA"
  content = "100::"
  proxied = true
  ttl     = 1
}

resource "cloudflare_dns_record" "a_records" {
  for_each = {
    "@"               = "75.2.60.5"
    "eng"             = "34.98.67.170"
    "gcsweb"          = "34.98.121.133"
    "monitoring.prow" = "35.227.227.231"
    "prow"            = "35.190.0.30"
    "prow-private"    = "35.190.0.30"
  }

  zone_id = var.zone_id
  name    = each.key
  ttl     = 300
  type    = "A"
  content = each.value
}

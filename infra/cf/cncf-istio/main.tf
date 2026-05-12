variable "account_id" {
  description = "Cloudflare account ID"
  type        = string
  default     = "1eb8152ceed73d3ee6f66c3557819d4c"
}

variable "zone_id" {
  description = "Cloudflare zone ID"
  type        = string
  default     = "e38050a1f3f682c6dbeba6087625fbe4"
}

locals {
  istio_registry_mappings = {
    release = "gcr.io/istio-release"
    testing = "gcr.io/istio-testing"
    prerelease-testing = "gcr.io/istio-prerelease-testing"
  }
}

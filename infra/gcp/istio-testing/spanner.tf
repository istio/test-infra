# Setup spanner instance for policy bot
# Policy bot stores a variety of state in spanner about github events, etc
resource "google_spanner_database" "main" {
  database_dialect         = "GOOGLE_STANDARD_SQL"
  instance                 = "istio-policy-bot"
  name                     = "main"
  project                  = "istio-testing"
  version_retention_period = "1h"
}
resource "google_spanner_instance" "istio_policy_bot" {
  config        = "projects/istio-testing/instanceConfigs/regional-us-west1"
  display_name  = "Istio Policy Bot"
  force_destroy = false
  name          = "istio-policy-bot"
  num_nodes     = 7
  project       = "istio-testing"
}
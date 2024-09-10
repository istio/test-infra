resource "google_project_service" "project" {
  project = local.project_number

  for_each = toset([
    "bigquery.googleapis.com",
    "bigquerystorage.googleapis.com",
    "cloudkms.googleapis.com",
    "compute.googleapis.com",
    "container.googleapis.com",
    "iam.googleapis.com",
    "logging.googleapis.com",
    "monitoring.googleapis.com",
    "pubsub.googleapis.com",
    "secretmanager.googleapis.com",
    "servicemanagement.googleapis.com",
    "serviceusage.googleapis.com",
    "stackdriver.googleapis.com",
  ])

  service = each.key
  timeouts {}

  // TODO: terraform wants to set this. I think its basically a NOP, but to keep the plan empty for now we ignore this.
  lifecycle {
    ignore_changes = [
      disable_on_destroy,
    ]
  }
}

# TODO: this role exists, but for some reason the terraform import is broken.
# We can probably just not use roles and inline the permissions now anyways once we move everything over to terraform.
#resource "google_project_iam_custom_role" "e2e_testing" {
#  description = "Created on: 2017-07-17"
#  permissions = [
#    "logging.exclusions.get",
#    "logging.exclusions.list",
#    "logging.logEntries.list",
#    "logging.logMetrics.get",
#    "logging.logMetrics.list",
#    "logging.logServiceIndexes.list",
#    "logging.logServices.list",
#    "logging.logs.list",
#    "logging.sinks.get",
#    "logging.sinks.list",
#    "logging.usage.get",
#    "stackdriver.projects.get",
#    "storage.objects.create",
#    "storage.objects.delete",
#    "storage.objects.get",
#    "storage.objects.list",
#    "storage.objects.update",
#  ]
#  project = local.project_id
#  role_id = "CustomRole"
#  stage   = "GA"
#  title   = "E2E Testing"
#}

## Misc Service Accounts
# Used by policy bot.
resource "google_service_account" "istio_policy_bot" {
  account_id   = "istio-policy-bot"
  display_name = "Istio Policy Bot"
  project      = "istio-testing"
}

## Prow Job Service Accounts ##
# Used with WI as the "prowjob-default-sa" service account. This is the default for jobs
resource "google_service_account" "istio_prow_test_job_default" {
  account_id   = "istio-prow-test-job-default"
  description  = "Default service account used by Istio Prow jobs"
  display_name = "istio-prow-test-job-default"
  project      = "istio-testing"
}
# Used with WI as the "prowjob-advanced-sa" service account. This is used for jobs that need elevated permissions
# Now obsolete; prowjob-advanced-sa is not used.
resource "google_service_account" "istio_prow_test_job" {
  account_id   = "istio-prow-test-job"
  display_name = "Istio Prow Test Job Service Account"
  project      = "istio-testing"
}
# Used for WI in test-infra trusted jobs
resource "google_service_account" "prow_deployer" {
  account_id   = "prow-deployer"
  description  = "Used to deploy to the clusters in the istio-testing and istio-prow-build projects."
  display_name = "Prow Self Deployer SA"
  project      = "istio-testing"
}
# Used for WI in test-infra trusted jobs
resource "google_service_account" "testgrid_updater" {
  account_id   = "testgrid-updater"
  description  = "Service account to upload istio's TestGrid info to cloud storage"
  display_name = "testgrid-updater"
  project      = "istio-testing"
}


## Prow Infra Service Accounts ##
# Used for WI for external secrets deployment
resource "google_service_account" "kubernetes_external_secrets_sa" {
  account_id   = "kubernetes-external-secrets-sa"
  display_name = "kubernetes-external-secrets-sa"
  project      = "istio-testing"
}
# External Secrets has project level IAM to secrets in istio-testing. Grant it access to one it needs in istio-prow-build as well.
resource "google_secret_manager_secret_iam_member" "member" {
  project   = "istio-prow-build"
  secret_id = "github_istio-testing_pusher"
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.kubernetes_external_secrets_sa.email}"
}

# Used for WI for prow control plane deployment
resource "google_service_account" "prow_control_plane" {
  account_id   = "prow-control-plane"
  description  = "Used by prow components"
  display_name = "prow-control-plane"
  project      = "istio-testing"
}
data "google_iam_policy" "prow_control_plane" {
  binding {
    role = "roles/iam.workloadIdentityUser"

    members = [
      "serviceAccount:istio-testing.svc.id.goog[default/crier]",
      "serviceAccount:istio-testing.svc.id.goog[default/deck]",
      "serviceAccount:istio-testing.svc.id.goog[default/deck-private]",
      "serviceAccount:istio-testing.svc.id.goog[default/hook]",
      "serviceAccount:istio-testing.svc.id.goog[default/prow-controller-manager]",
      "serviceAccount:istio-testing.svc.id.goog[default/sinker]",
      "serviceAccount:istio-testing.svc.id.goog[default/tide]",
    ]
  }
}
resource "google_service_account_iam_policy" "prow_control_plane" {
  service_account_id = google_service_account.prow_control_plane.name
  policy_data        = data.google_iam_policy.prow_control_plane.policy_data
}

module "prowjob_bots_deployer_account" {
  source            = "../modules/workload-identity-service-account"
  project_id        = local.project_id
  name              = "prowjob-bots-deployer"
  description       = "ProwJob SA for deploying istio/bots"
  cluster_namespace = local.pod_namespace
  # Grant container admin so we can get credentials to deploy to the policy-bot GKE cluster
  project_roles = [
    { role = "roles/container.admin" },
  ]
  gcs_acls = [
    { bucket = "artifacts.istio-testing.appspot.com", role = "OWNER" },
  ]
  prowjob = true
}

resource "google_project_iam_member" "project-admins" {
  for_each = toset(local.terraform_infra_admins)
  project  = local.project_id
  role     = "roles/resourcemanager.projectIamAdmin"
  member   = "user:${each.key}"
}

resource "google_project_iam_member" "owners" {
  for_each = toset(local.terraform_infra_admins)
  project  = local.project_id
  role     = "roles/owners"
  member   = "user:${each.key}"
}


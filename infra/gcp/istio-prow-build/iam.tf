# WARNING: the description here is false. I have no clue what this is used for TBH
# The default prowjob SA is istio-prow-test-job-default@istio-testing.iam.gserviceaccount.com
resource "google_service_account" "istio_prow_jobs" {
  account_id   = "istio-prow-jobs"
  description  = "The default service account that will be used for Prow job workloads."
  display_name = "istio-prow-jobs"
  project      = "istio-prow-build"
}

# ProwJob SA used for release jobs. This is the most privileged service account, and should be used only on trusted code
# with extreme caution.
# Do not use this for other purposes! Create a new, more scoped, account.
# This is granted KMS access in keys.tf
module "prowjob_release_account" {
  source            = "../modules/workload-identity-service-account"
  project_id        = local.project_id
  name              = "prowjob-release"
  description       = "Service account used for prow release jobs. Highly privileged."
  cluster_namespace = local.pod_namespace
  secrets = [
    { name = "release_docker_istio" },
    { name = "release_github_istio-release" },
    { name = "release_grafana_istio" },
    # This may look a bit weird, but for building base images we need dockerhub + istio-testing account, not istio-release account.
    # So we give this prow job access.
    # We could scope this down a bit with 2 accounts, but all of these are super high privilege anyways so it is simpler to keep them together.
    { name = "github_istio-testing_pusher" },
  ]
  gcs_acls = [
    { bucket = "istio-prerelease", role = "OWNER" },
    { bucket = "istio-release", role = "OWNER" },
    { bucket = "artifacts.istio-release.appspot.com", role = "OWNER" },
    { bucket = "artifacts.istio-prerelease-testing.appspot.com", role = "OWNER" },
  ]
  prowjob = true
}

# ProwJob SA used for jobs requiring RBE access.
module "prowjob_rbe_account" {
  source            = "../modules/workload-identity-service-account"
  project_id        = local.project_id
  name              = "prowjob-rbe"
  description       = "Service account used for prow jobs requiring RBE access (istio/proxy)."
  cluster_namespace = local.pod_namespace
  project_roles = [
    { role = "roles/remotebuildexecution.actionCacheWriter", project = "istio-testing" },
    { role = "roles/remotebuildexecution.artifactCreator", project = "istio-testing" },
  ]
  prowjob = true
}

# ProwJob SA used for jobs requiring GitHub API readonly access.
# This is granted secret access in secrets.tf
module "prowjob_github_read_account" {
  source            = "../modules/workload-identity-service-account"
  project_id        = local.project_id
  name              = "prowjob-github-read"
  description       = "Service account used for prow jobs requiring GitHub read access."
  cluster_namespace = local.pod_namespace
  secrets = [
    { name = "github-read_github_read" },
  ]
  prowjob = true
}

# Service account that has permissions for GitHub from istio-testing account. Has permissions to push PRs
module "prowjob_github_istio_testing_account" {
  source            = "../modules/workload-identity-service-account"
  project_id        = local.project_id
  name              = "prowjob-github-istio-testing"
  description       = "Service account that has permissions for GitHub from istio-testing account. Has permissions to push PRs."
  cluster_namespace = local.pod_namespace
  secrets = [
    { name = "github_istio-testing_pusher" },
  ]
  prowjob = true
}

module "prowjob_build_tools_account" {
  source            = "../modules/workload-identity-service-account"
  project_id        = local.project_id
  name              = "prowjob-build-tools"
  description       = "Service account that has permissions to push to gcr.io/istio-testing and to push PRs as istio-testing account."
  cluster_namespace = local.pod_namespace
  secrets = [
    { name = "github_istio-testing_pusher" },
  ]
  gcs_acls = [
    { bucket = "artifacts.istio-testing.appspot.com", role = "OWNER" },
  ]
  prowjob = true
}

module "prowjob_testing_write_account" {
  source            = "../modules/workload-identity-service-account"
  project_id        = local.project_id
  name              = "prowjob-testing-write"
  description       = "Service account that has permissions to push to gcr.io/istio-testing and gs://istio-build."
  cluster_namespace = local.pod_namespace
  gcs_acls = [
    { bucket = "artifacts.istio-testing.appspot.com", role = "OWNER" },
    { bucket = "istio-build", role = "OWNER" },
  ]
  project_roles = [
    { role = "roles/remotebuildexecution.actionCacheWriter", project = "istio-testing" },
    { role = "roles/remotebuildexecution.artifactCreator", project = "istio-testing" },
  ]
  # Allow the same SA in istio-testing to access the secret. This is the trusted build cluster.
  additional_workload_identity_principals = [
    "serviceAccount:istio-testing.svc.id.goog[test-pods/prowjob-testing-write]"
  ]
  prowjob = true
}

module "opentelemetry_collector_account" {
  source            = "../modules/workload-identity-service-account"
  project_id        = local.project_id
  name              = "opentelemetry-collector"
  description       = "Service account that has permissions to telemetry data."
  cluster_namespace = "opentelemetry"
  project_roles = [
    { role = "roles/cloudtrace.agent", project = local.project_id },
  ]
  prowjob = false
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
  role     = "roles/owner"
  member   = "user:${each.key}"
}

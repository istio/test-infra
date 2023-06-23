resource "google_service_account" "external_secrets_private" {
  account_id   = "external-secrets-private"
  description  = "Kubernetes External Secrets controller for private cluster"
  display_name = "external-secrets-private"
  project      = "istio-prow-build"
}
# WARNING: the description here is false. I have no clue what this is used for TBH
# The default prowjob SA is istio-prow-test-job-default@istio-testing.iam.gserviceaccount.com
resource "google_service_account" "istio_prow_jobs" {
  account_id   = "istio-prow-jobs"
  description  = "The default service account that will be used for Prow job workloads."
  display_name = "istio-prow-jobs"
  project      = "istio-prow-build"
}
resource "google_service_account" "kubernetes_external_secrets_sa" {
  account_id   = "kubernetes-external-secrets-sa"
  description  = "Service account used by external secrets controller"
  display_name = "kubernetes-external-secrets-sa"
  project      = "istio-prow-build"
}
resource "google_service_account" "prow_internal_storage" {
  account_id   = "prow-internal-storage"
  description  = "Internal Prow SA for istio-private-build GCS. "
  display_name = "Prow Internal Storage"
  project      = "istio-prow-build"
}

# ProwJob SA used for release jobs. This is the most privileged service account, and should be used only on trusted code
# with extreme caution.
# Do not use this for other purposes! Create a new, more scoped, account.
# This is granted secret access in secrets.tf
module "prowjob_release_account" {
  source            = "../modules/workload-identity-service-account"
  project_id        = local.project_id
  name              = "prowjob-release"
  description       = "Service account used for prow release jobs. Highly privileged."
  cluster_namespace = local.pod_namespace
  prowjob           = true
}

# ProwJob SA used for jobs requiring RBE access.
module "prowjob_rbe_account" {
  source            = "../modules/workload-identity-service-account"
  project_id        = local.project_id
  name              = "prowjob-rbe"
  description       = "Service account used for prow jobs requireing RBE access (istio/proxy)."
  cluster_namespace = local.pod_namespace
  project_roles = [
    { role = "roles/remotebuildexecution.actionCacheWriter", project = "istio-testing" },
    { role = "roles/remotebuildexecution.artifactCreator", project = "istio-testing" },
  ]
  prowjob = true
}

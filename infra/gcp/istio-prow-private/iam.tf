# Because private project is only for trusted individuals, we do not use fine-scoped service accounts.
# Instead, everything runs as one service account with access to what we need.
# Granted AR access in ar.tf.
# RBE is not used in private cluster, so no permission needed there.
# DockerHub, Grafana, and GitHub tokens are not used, so no secret access needed.
module "prowjob_private_account" {
  source            = "../modules/workload-identity-service-account"
  project_id        = local.project_id
  name              = "prowjob-private"
  description       = "Service account that has permissions for all private jobs."
  cluster_namespace = local.pod_namespace
  gcs_iam = [
    { bucket = google_storage_bucket.istio_build_private.name, role = "roles/storage.objectAdmin" },
    { bucket = google_storage_bucket.istio_prerelease_private.name, role = "roles/storage.objectAdmin" },
  ]
  prowjob = true
  prowjob-bucket = "istio-prow-private"
}

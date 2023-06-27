/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// creates a service account in project_id with name and description
// usable by pods in cluster_project_id
// running in namespace cluster_namespace
// running as cluster_serviceaccount_name

locals {
  description                 = var.description != "" ? var.description : var.name
  display_name                = var.display_name != "" ? var.display_name : var.name
  cluster_project_id          = var.cluster_project_id != "" ? var.cluster_project_id : var.project_id
  cluster_serviceaccount_name = var.cluster_serviceaccount_name != "" ? var.cluster_serviceaccount_name : var.name
}

resource "google_service_account" "serviceaccount" {
  project      = var.project_id
  account_id   = var.name
  display_name = local.display_name
  description  = local.description
}
data "google_iam_policy" "workload_identity" {
  binding {
    members = [
      "serviceAccount:${local.cluster_project_id}.svc.id.goog[${var.cluster_namespace}/${local.cluster_serviceaccount_name}]"
    ]
    role = "roles/iam.workloadIdentityUser"
  }
}
// authoritative binding, replaces any existing IAM policy on the service account
resource "google_service_account_iam_policy" "serviceaccount_iam" {
  service_account_id = google_service_account.serviceaccount.name
  policy_data        = data.google_iam_policy.workload_identity.policy_data
}
// optional: roles to grant the serviceaccount on the project
resource "google_project_iam_member" "project_roles" {
  for_each = { for k, v in var.project_roles : k => v }
  project  = coalesce(each.value.project, var.project_id)
  role     = each.value.role
  member   = "serviceAccount:${google_service_account.serviceaccount.email}"
}
// optional: GCS ACLs to grant the serviceaccount on the project
resource "google_storage_bucket_access_control" "acls" {
  for_each = { for k, v in var.gcs_acls : k => v }
  bucket   = each.value.bucket
  role     = each.value.role
  entity   = "user-${google_service_account.serviceaccount.email}"
}
// optional: GCS IAM grants access to the serviceaccount on the project
resource "google_storage_bucket_iam_member" "gcs_member" {
  for_each = { for k, v in var.gcs_acls : k => v }
  bucket   = each.value.bucket
  role     = each.value.role
  member   = "serviceAccount:${google_service_account.serviceaccount.email}"
}

// optional: secrets to grant access to the serviceaccount on the project
resource "google_secret_manager_secret_iam_member" "member" {
  for_each  = { for k, v in var.secrets : k => v }
  project   = coalesce(each.value.project, var.project_id)
  secret_id = each.value.name
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.serviceaccount.email}"
}
// If this is going to be used for prowjobs, then we need to give it access to gs://istio-prow to write logs.
resource "google_storage_bucket_iam_member" "member" {
  count  = var.prowjob ? 1 : 0
  bucket = var.bucket
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.serviceaccount.email}"
}

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

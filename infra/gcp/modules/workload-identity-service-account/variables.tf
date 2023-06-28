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

variable "project_id" {
  description = "The id of the project hosting the serviceaccount, eg: my-awesome-project"
  type        = string
}

variable "name" {
  description = "The name of the serviceaccount, eg: my-awesome-sa"
  type        = string
}

variable "description" {
  description = "The description of the service account, eg: My Awesome Service Account (default: name)"
  type        = string
  default     = ""
}

variable "display_name" {
  description = "The display name of the service account"
  type        = string
  default     = ""
}

variable "cluster_project_id" {
  description = "The id of the project hosting clusters that will use the serviceaccount, eg: my-awesome-cluster-project (default: project_id)"
  type        = string
  default     = ""
}

variable "cluster_serviceaccount_name" {
  description = "The name of the kubernetes service account that will bind to the service account, eg: my-cluster-sa (default: name)"
  type        = string
  default     = ""
}

variable "cluster_namespace" {
  description = "The namespace of the kubernetes service account that will bind to the service account, eg: my-namespace"
  type        = string
}

variable "project_roles" {
  description = "A list of roles to bind to the serviceaccount in its project"
  type = list(object({
    role    = string
    project = optional(string)
  }))
  default = []
}

variable "secrets" {
  description = "A list of secrets to give access to the serviceaccount in its project"
  type = list(object({
    name    = string
    project = optional(string)
  }))
  default = []
}

variable "gcs_acls" {
  description = "A list of buckets to add ACLs for. Note: prefer using IAM for GCS; this is for legacy bucket configurations"
  type = list(object({
    bucket = string
    role   = string
  }))
  default = []
}

variable "gcs_iam" {
  description = "A list of buckets to add IAM bindings for. Use gcs_acls for the legacy ACL config"
  type = list(object({
    bucket = string
    role   = string
  }))
  default = []
}

variable "prowjob" {
  description = "Set to true if this service account will be used for prowjobs"
  type        = bool
}

variable "prowjob_bucket" {
  description = "If 'prowjob' is true, which bucket to grant access to"
  type        = string
  default     = "istio-prow"
}

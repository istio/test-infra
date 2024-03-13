/*
Copyright 2020 The Kubernetes Authors.

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

resource "google_container_node_pool" "node_pool" {
  name = var.name

  project  = var.project_name
  location = var.location
  cluster  = var.cluster_name

  // Auto repair, and auto upgrade nodes to match the master version
  management {
    auto_repair  = true
    auto_upgrade = true
  }

  // Autoscale the cluster as needed. Note if using a regional cluster these values will be multiplied by 3
  initial_node_count = var.initial_count
  autoscaling {
    min_node_count = var.min_count
    max_node_count = var.max_count
  }

  // Set machine type, and enable all oauth scopes tied to the service account
  node_config {
    image_type   = "COS_CONTAINERD"
    machine_type = var.machine_type
    disk_size_gb = var.disk_size_gb
    disk_type    = var.disk_type
    labels       = var.labels
    spot         = var.spot

    service_account = var.service_account
    oauth_scopes    = ["https://www.googleapis.com/auth/cloud-platform"]

    // Needed for workload identity
    workload_metadata_config {
      mode = "GKE_METADATA"
    }
    metadata = {
      disable-legacy-endpoints = "true"
    }

    linux_node_config {
      // We cannot currently run prow on Cgroupsv2
      cgroup_mode = "CGROUP_MODE_V1"
    }
  }


  // If we need to destroy the node pool, create the new one before destroying
  // the old one
  lifecycle {
    create_before_destroy = true
    ignore_changes = [
      # https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/container_node_pool#initial_node_count
      initial_node_count,
      # https://www.terraform.io/docs/providers/google/r/container_cluster.html#taint
      node_config[0].taint,
    ]
  }
}

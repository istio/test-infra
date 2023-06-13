# The 'prow' cluster hosts the control plane for Prow, and runs a few trusted jobs
# The majority of jobs are run in a different project, `istio-prow-build`.
resource "google_container_cluster" "prow" {
  addons_config {
    http_load_balancing {
      disabled = false
    }

    network_policy_config {
      disabled = true
    }
  }

  cluster_autoscaling {
    enabled = false
  }

  cluster_ipv4_cidr = "10.44.0.0/14"

  enable_legacy_abac = true
  location           = "us-west1-a"

  logging_config {
    enable_components = ["SYSTEM_COMPONENTS", "WORKLOADS"]
  }

  master_auth {
    client_certificate_config {
      issue_client_certificate = true
    }
  }

  monitoring_config {
    enable_components = ["SYSTEM_COMPONENTS"]

    managed_prometheus {
      enabled = true
    }
  }

  name    = "prow"
  network = "projects/istio-testing/global/networks/default"

  network_policy {
    enabled  = false
    provider = "CALICO"
  }

  networking_mode = "ROUTES"

  node_version = "1.24.12-gke.1000"

  private_cluster_config {
    enable_private_endpoint = false

    master_global_access_config {
      enabled = false
    }
  }

  project = "istio-testing"

  release_channel {
    channel = "REGULAR"
  }

  subnetwork = "projects/istio-testing/regions/us-west1/subnetworks/default"

  workload_identity_config {
    workload_pool = "istio-testing.svc.id.goog"
  }

  # Cluster is so old this doesn't exist in our cluster API.
  # Probably we can explicit set this to false to make things normal, for now ignore it.
  lifecycle {
    ignore_changes = [enable_shielded_nodes, timeouts]
  }
}

# Prow cluster just uses one node pool
resource "google_container_node_pool" "prow_pool" {
  autoscaling {
    max_node_count = 8
    min_node_count = 4
  }

  cluster            = "prow"
  initial_node_count = 5
  location           = "us-west1-a"

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  name = "prod-node-pool"

  node_config {
    disk_size_gb = 100
    disk_type    = "pd-standard"
    image_type   = "COS_CONTAINERD"

    labels = {
      prod = "prow"
    }

    machine_type = "e2-standard-4"

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes    = ["https://www.googleapis.com/auth/devstorage.read_only", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/service.management.readonly", "https://www.googleapis.com/auth/servicecontrol", "https://www.googleapis.com/auth/trace.append"]
    service_account = "default"

    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }

  node_count     = 4
  node_locations = ["us-west1-a"]
  project        = "istio-testing"

  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }

  version = "1.24.12-gke.1000"
}
# policy_bot hosts a deployment of https://github.com/istio/bots
resource "google_container_cluster" "policy_bot" {
  addons_config {
    horizontal_pod_autoscaling {
      disabled = false
    }

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

  cluster_ipv4_cidr = "10.48.0.0/14"

  database_encryption {
    state = "DECRYPTED"
  }

  default_max_pods_per_node = 110
  enable_shielded_nodes     = false
  location                  = "us-central1-a"

  logging_config {
    enable_components = ["SYSTEM_COMPONENTS", "WORKLOADS"]
  }

  master_auth {
    client_certificate_config {
      issue_client_certificate = false
    }
  }

  monitoring_config {
    enable_components = ["SYSTEM_COMPONENTS"]
  }

  name    = "policy-bot"
  network = "projects/istio-testing/global/networks/default"

  network_policy {
    enabled = false
  }

  networking_mode = "ROUTES"

  node_version = "1.23.17-gke.2000"

  project = "istio-testing"

  release_channel {
    channel = "UNSPECIFIED"
  }

  subnetwork = "projects/istio-testing/regions/us-central1/subnetworks/default"
}
# Single node pool
resource "google_container_node_pool" "policy_bot_pool" {
  autoscaling {
    max_node_count = 10
    min_node_count = 3
  }

  cluster            = "policy-bot"
  initial_node_count = 3
  location           = "us-central1-a"

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  name = "default-pool"

  node_config {
    disk_size_gb = 100
    disk_type    = "pd-standard"
    image_type   = "COS"
    machine_type = "n1-standard-1"

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes    = ["https://www.googleapis.com/auth/devstorage.read_only", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/service.management.readonly", "https://www.googleapis.com/auth/servicecontrol", "https://www.googleapis.com/auth/trace.append"]
    service_account = "default"

    shielded_instance_config {
      enable_integrity_monitoring = true
    }
  }

  node_count     = 3
  node_locations = ["us-central1-a"]
  project        = "istio-testing"

  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }

  version = "1.23.17-gke.2000"
}
# terraform import google_container_node_pool.default_pool istio-testing/us-central1-a/policy-bot/default-pool
# Prow private cluster is the private mirror of the "prow" cluster.
# Access to this is restricted to only private infrastructure.
resource "google_container_cluster" "prow_private" {
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

  database_encryption {
    state = "DECRYPTED"
  }

  default_max_pods_per_node = 110
  enable_shielded_nodes     = false

  ip_allocation_policy {
    cluster_ipv4_cidr_block  = "10.4.0.0/14"
    services_ipv4_cidr_block = "10.2.0.0/20"
  }

  location = "us-west1-a"

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

  name    = "prow-private"
  network = "projects/istio-prow-build/global/networks/default"

  network_policy {
    enabled = false
  }

  networking_mode = "VPC_NATIVE"

  node_version = "1.24.12-gke.500"

  private_cluster_config {
    enable_private_endpoint = false

    master_global_access_config {
      enabled = false
    }
  }

  project = "istio-prow-build"

  release_channel {
    channel = "REGULAR"
  }

  resource_labels = {
    role  = "prow-private"
    owner = "oss-istio"
  }

  subnetwork = "projects/istio-prow-build/regions/us-west1/subnetworks/default"

  workload_identity_config {
    workload_pool = "istio-prow-build.svc.id.goog"
  }
}

# Mirrors 'prow_build'
resource "google_container_node_pool" "prow_private_build" {
  autoscaling {
    max_node_count = 10
    min_node_count = 0
  }

  cluster            = "prow-private"
  initial_node_count = 2
  location           = "us-west1-a"

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  max_pods_per_node = 110
  name              = "istio-build-pool-containerd"

  network_config {
    pod_ipv4_cidr_block = "10.4.0.0/14"
    pod_range           = "gke-prow-private-pods-07555fa6"
  }

  node_config {
    disk_size_gb = 2000
    disk_type    = "pd-ssd"
    image_type   = "COS_CONTAINERD"

    labels = {
      testing = "build-pool"
    }

    machine_type = "n1-highmem-64"

    metadata = {
      testing                  = "build-pool"
      disable-legacy-endpoints = "true"
    }

    oauth_scopes    = ["https://www.googleapis.com/auth/devstorage.read_only", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/service.management.readonly", "https://www.googleapis.com/auth/servicecontrol", "https://www.googleapis.com/auth/trace.append"]
    service_account = "default"

    shielded_instance_config {
      enable_integrity_monitoring = true
    }

    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }

  node_count     = 2
  node_locations = ["us-west1-a"]
  project        = "istio-prow-build"

  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }

  version = "1.24.12-gke.500"
}

# Mirrors 'prow_test'
resource "google_container_node_pool" "prow_private_test" {
  autoscaling {
    max_node_count = 15
    min_node_count = 0
  }

  cluster            = "prow-private"
  initial_node_count = 0
  location           = "us-west1-a"

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  max_pods_per_node = 110
  name              = "istio-test-pool"

  network_config {
    pod_ipv4_cidr_block = "10.4.0.0/14"
    pod_range           = "gke-prow-private-pods-07555fa6"
  }

  node_config {
    disk_size_gb = 256
    disk_type    = "pd-ssd"
    image_type   = "COS_CONTAINERD"

    labels = {
      testing = "test-pool"
    }

    machine_type = "e2-standard-16"

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes    = ["https://www.googleapis.com/auth/cloud-platform"]
    service_account = "istio-prow-jobs@istio-prow-build.iam.gserviceaccount.com"

    shielded_instance_config {
      enable_integrity_monitoring = true
    }

    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }

  node_count     = 0
  node_locations = ["us-west1-a"]
  project        = "istio-prow-build"

  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }

  version = "1.24.12-gke.500"
}

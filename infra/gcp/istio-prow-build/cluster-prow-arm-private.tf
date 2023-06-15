# Prow private cluster is the private mirror of the "prow-arm" cluster.
# Access to this is restricted to only private infrastructure.
resource "google_container_cluster" "prow_arm_private" {
  addons_config {
    dns_cache_config {
      enabled = false
    }

    gce_persistent_disk_csi_driver_config {
      enabled = true
    }

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

  binary_authorization {
    evaluation_mode = "DISABLED"
  }

  cluster_autoscaling {
    enabled = false
  }

  database_encryption {
    state = "DECRYPTED"
  }

  datapath_provider         = "LEGACY_DATAPATH"
  default_max_pods_per_node = 110

  default_snat_status {
    disabled = false
  }

  enable_shielded_nodes = true

  ip_allocation_policy {
    cluster_ipv4_cidr_block  = "10.52.0.0/14"
    services_ipv4_cidr_block = "10.56.0.0/20"
  }

  location = "us-central1-f"

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

    managed_prometheus {
      enabled = false
    }
  }

  name    = "prow-arm-private"
  network = "projects/istio-prow-build/global/networks/default"

  network_policy {
    enabled  = false
    provider = "PROVIDER_UNSPECIFIED"
  }

  networking_mode = "VPC_NATIVE"

  node_version = "1.25.8-gke.500"

  notification_config {
    pubsub {
      enabled = false
    }
  }

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

  subnetwork = "projects/istio-prow-build/regions/us-central1/subnetworks/default"

  workload_identity_config {
    workload_pool = "istio-prow-build.svc.id.goog"
  }
}

# TODO: we have "prow_arm_build" but no "prow_arm_private_build"
# As noted in the prow_arm_build node pool, I think its not used.

# Mirror of 'prow_arm_default'
resource "google_container_node_pool" "prow_arm_private_default" {
  autoscaling {
    total_max_node_count = 5
    total_min_node_count = 0
  }

  cluster            = "prow-arm-private"
  initial_node_count = 1
  location           = "us-central1-f"

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  max_pods_per_node = 110
  name              = "default-pool"

  network_config {
    pod_ipv4_cidr_block = "10.52.0.0/14"
    pod_range           = "gke-prow-arm-private-pods-99952dee"
  }

  node_config {
    disk_size_gb = 256
    disk_type    = "pd-ssd"
    image_type   = "COS_CONTAINERD"
    machine_type = "n1-standard-8"

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes    = ["https://www.googleapis.com/auth/devstorage.read_only", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/service.management.readonly", "https://www.googleapis.com/auth/servicecontrol", "https://www.googleapis.com/auth/trace.append"]
    service_account = "default"

    shielded_instance_config {
      enable_integrity_monitoring = true
    }

    spot = true

    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }

  node_count     = 1
  node_locations = ["us-central1-f"]
  project        = "istio-prow-build"

  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }

  version = "1.25.8-gke.500"
}

# Mirror of 'prow_arm_test_spot'
resource "google_container_node_pool" "prow_arm_private_test_spot" {
  cluster            = "prow-arm-private"
  initial_node_count = 3
  location           = "us-central1-f"

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  max_pods_per_node = 110
  name              = "t2a-spot"

  network_config {
    pod_ipv4_cidr_block = "10.52.0.0/14"
    pod_range           = "gke-prow-arm-private-pods-99952dee"
  }

  node_config {
    disk_size_gb = 256
    disk_type    = "pd-ssd"
    image_type   = "COS_CONTAINERD"

    labels = {
      testing = "test-pool"
    }

    machine_type = "t2a-standard-16"

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes    = ["https://www.googleapis.com/auth/devstorage.read_only", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/service.management.readonly", "https://www.googleapis.com/auth/servicecontrol", "https://www.googleapis.com/auth/trace.append"]
    service_account = "default"

    shielded_instance_config {
      enable_integrity_monitoring = true
    }

    spot = true

    taint {
      effect = "NO_SCHEDULE"
      key    = "kubernetes.io/arch"
      value  = "arm64"
    }

    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }

  node_count     = 3
  node_locations = ["us-central1-f"]
  project        = "istio-prow-build"

  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }

  version = "1.25.8-gke.500"

  # ARM defaults to cgroups v2. However, our test setup (kind) do not yet support this
  # Terraform does not yet support this mode, so we have to just set it manually and ignore changes
  # TODO: https://github.com/hashicorp/terraform-provider-google/issues/12712
  #    linux_node_config {
  #      cgroup_mode = "CGROUP_MODE_V1"
  #    }
  lifecycle {
    ignore_changes = [
      node_config[0].linux_node_config,
    ]
  }
}

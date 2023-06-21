# prow_arm provides a cluster that hosts ARM jobs.
# This mirrors the "prow" cluster, but with ARM.
# Ideally, this could just be the "prow" cluster with an ARM node pool.
# However, ARM nodes availability is limited to us-central1-f, so we have our own.
# In the future this could be consolidated.
resource "google_container_cluster" "prow_arm" {
  addons_config {
    gce_persistent_disk_csi_driver_config {
      enabled = true
    }

    network_policy_config {
      disabled = true
    }
  }

  cluster_autoscaling {
    enabled = false
  }

  cluster_ipv4_cidr = "10.12.0.0/14"

  database_encryption {
    state = "DECRYPTED"
  }

  enable_shielded_nodes = true
  location              = "us-central1-f"

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
  }

  name    = "prow-arm"
  network = "projects/istio-prow-build/global/networks/default"

  network_policy {
    enabled  = false
    provider = "PROVIDER_UNSPECIFIED"
  }

  networking_mode = "ROUTES"

  node_version = "1.26.5-gke.1200"

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
    channel = "RAPID"
  }

  resource_labels = {
    role  = "prow"
    owner = "oss-istio"
  }

  subnetwork = "projects/istio-prow-build/regions/us-central1/subnetworks/default"

  workload_identity_config {
    workload_pool = "istio-prow-build.svc.id.goog"
  }
}

# The "build" cluster doesn't run anything as far as I can tell. This can probably be removed.
resource "google_container_node_pool" "prow_arm_build" {
  autoscaling {
    max_node_count = 1
    min_node_count = 1
  }

  cluster            = "prow-arm"
  initial_node_count = 0
  location           = "us-central1-f"

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  name = "arm-build-pool-large"

  node_config {
    disk_size_gb = 100
    disk_type    = "pd-balanced"

    gvnic {
      enabled = true
    }

    image_type   = "COS_CONTAINERD"
    machine_type = "t2a-standard-16"

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes    = ["https://www.googleapis.com/auth/devstorage.read_only", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/service.management.readonly", "https://www.googleapis.com/auth/servicecontrol", "https://www.googleapis.com/auth/trace.append"]
    service_account = "default"

    shielded_instance_config {
      enable_integrity_monitoring = true
    }

    taint {
      effect = "NO_SCHEDULE"
      key    = "kubernetes.io/arch"
      value  = "arm64"
    }

    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }

  node_count     = 1
  node_locations = ["us-central1-f"]
  project        = "istio-prow-build"

  upgrade_settings {
    max_surge       = 1
    max_unavailable = 1
  }

  version = "1.26.5-gke.1200"
}

# The default pool hosts an x86 node pool. This is to run some of the prow infrastructure which isn't arm compatible.
# This is just a single node without scaling, no tests run here.
resource "google_container_node_pool" "prow_arm_default" {
  cluster            = "prow-arm"
  initial_node_count = 1
  location           = "us-central1-f"

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  name = "default-pool"

  node_config {
    disk_size_gb = 100
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

  version = "1.26.5-gke.1200"
}

# This pool provides the actual ARM (t2a) instances for tests.
# Spot instances are used as quota is capped for ARM nodes, and its cheaper.
# Currently, autoscaling is disabled due to ongoing networking issues on ARM.
resource "google_container_node_pool" "prow_arm_test_spot" {
  autoscaling {
    total_max_node_count = 6
    total_min_node_count = 6
  }

  cluster  = "prow-arm"
  location = "us-central1-f"

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  name = "t2a-spot"

  node_config {
    disk_size_gb = 256
    disk_type    = "pd-ssd"

    gvnic {
      enabled = true
    }

    image_type = "COS_CONTAINERD"

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

  node_count     = 6
  node_locations = ["us-central1-f"]
  project        = "istio-prow-build"

  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }

  version = "1.26.5-gke.1200"

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

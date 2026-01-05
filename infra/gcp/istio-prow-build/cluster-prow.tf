# Prow cluster is the core build cluster. Most jobs end up running here
resource "google_container_cluster" "prow" {
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
    cluster_ipv4_cidr_block  = "10.44.0.0/14"
    services_ipv4_cidr_block = "10.0.0.0/20"
  }

  location = "us-west1-a"

  logging_config {
    enable_components = ["SYSTEM_COMPONENTS", "WORKLOADS"]
  }

  maintenance_policy {
    recurring_window {
      end_time   = "2021-01-12T08:00:00Z"
      recurrence = "FREQ=WEEKLY;BYDAY=SA,SU"
      start_time = "2021-01-11T08:00:00Z"
    }
  }

  master_auth {
    client_certificate_config {
      issue_client_certificate = false
    }
  }

  monitoring_config {
    enable_components = ["SYSTEM_COMPONENTS"]
  }

  name    = "prow"
  network = "projects/istio-prow-build/global/networks/default"

  network_policy {
    enabled = false
  }

  networking_mode = "VPC_NATIVE"

  project = "istio-prow-build"

  release_channel {
    channel = "REGULAR"
  }

  resource_labels = {
    role  = "prow"
    owner = "oss-istio"
  }

  subnetwork = "projects/istio-prow-build/regions/us-west1/subnetworks/default"

  workload_identity_config {
    workload_pool = "istio-prow-build.svc.id.goog"
  }
}

# Prow 'build' node pool is used for large jobs, mostly istio/proxy.
# Consists of 64 core machines.
# Note: despite the naming "build" vs "test", a lot of build jobs use the test pool.
resource "google_container_node_pool" "prow_build" {
  autoscaling {
    max_node_count = 10
    min_node_count = 0
  }

  cluster            = "prow"
  initial_node_count = 2
  location           = "us-west1-a"

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  max_pods_per_node = 110
  name              = "istio-build-pool-containerd-n2"

  network_config {
    pod_ipv4_cidr_block = "10.44.0.0/14"
    pod_range           = "gke-prow-pods-477396f0"
  }

  node_config {
    disk_size_gb = 2000
    disk_type    = "pd-ssd"
    image_type   = "COS_CONTAINERD"

    labels = {
      testing = "build-pool"
    }

    machine_type = "n1-standard-64"

    metadata = {
      disable-legacy-endpoints = "true"
      testing                  = "build-pool"
    }

    oauth_scopes = [
      "https://www.googleapis.com/auth/devstorage.read_only", "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/service.management.readonly",
      "https://www.googleapis.com/auth/servicecontrol", "https://www.googleapis.com/auth/trace.append"
    ]
    service_account = "default"

    shielded_instance_config {
      enable_integrity_monitoring = true
    }

    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }

  node_locations = ["us-west1-a"]
  project        = "istio-prow-build"

  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }
}


# Prow 'test' node pool is used for most jobs
# Consists of e2-16 nodes.
# Note: despite the naming "build" vs "test", a lot of build jobs use the test pool.
module "prow_test" {
  source = "../modules/gke-nodepool"

  name         = "istio-test-pool-e2"
  cluster_name = "prow"
  location     = "us-west1-a"

  machine_type  = "e2-standard-16"
  initial_count = 2
  max_count     = 60
  min_count     = 2

  disk_size_gb = 256
  disk_type    = "pd-ssd"

  project_name    = local.project_id
  service_account = "istio-prow-jobs@istio-prow-build.iam.gserviceaccount.com"

  labels = {
    "testing" = "test-pool"
  }
}

# Prow 'test' node pool is used for most jobs
# Consists of c3-22 nodes.
# Note: despite the naming "build" vs "test", a lot of build jobs use the test pool.
module "prow_test_c3" {
  source = "../modules/gke-nodepool"

  name         = "istio-test-pool-c3"
  cluster_name = "prow"
  location     = "us-west1-a"

  machine_type  = "c3-standard-22"
  initial_count = 0
  max_count     = 0
  min_count     = 0

  # C3 are a bit expensive, so drop cut price by using spot instances
  spot = true

  disk_size_gb = 256
  disk_type    = "pd-ssd"

  project_name    = local.project_id
  service_account = "istio-prow-jobs@istio-prow-build.iam.gserviceaccount.com"

  labels = {
    "testing" = "test-pool"
  }
}

# Prow 'test' node pool is used for most jobs
# Consists of n4-16 nodes.
# Note: despite the naming "build" vs "test", a lot of build jobs use the test pool.
module "prow_test_n4" {
  source = "../modules/gke-nodepool"

  name         = "istio-test-pool-n4"
  cluster_name = "prow"
  location     = "us-west1-a"

  machine_type  = "n4-standard-16"
  initial_count = 1
  max_count     = 1
  min_count     = 0

  # N4 are a bit expensive, so drop cut price by using spot instances
  spot = true

  disk_size_gb = 256
  disk_type    = "hyperdisk-balanced"

  project_name    = local.project_id
  service_account = "istio-prow-jobs@istio-prow-build.iam.gserviceaccount.com"

  labels = {
    "testing" = "test-pool"
  }
}

# Prow 'test' node pool is used for most jobs
# Consists of c4d-16 nodes.
# Note: despite the naming "build" vs "test", a lot of build jobs use the test pool.
module "prow_test_c4d" {
  source = "../modules/gke-nodepool"

  name         = "istio-test-pool-c4d"
  cluster_name = "prow"
  location     = "us-west1-a"

  machine_type  = "c4d-standard-16"
  initial_count = 1
  max_count     = 20
  min_count     = 1

  # C4D are a bit expensive, so drop cut price by using spot instances
  spot = true

  disk_size_gb = 256
  disk_type    = "hyperdisk-balanced"

  project_name    = local.project_id
  service_account = "istio-prow-jobs@istio-prow-build.iam.gserviceaccount.com"

  labels = {
    "testing" = "test-pool"
  }
}

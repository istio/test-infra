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

# The default pool hosts an x86 node pool. This is to run some of the prow infrastructure which isn't arm compatible.
# This is just a single node without scaling, no tests run here.
module "prow_arm_default" {
  source       = "../modules/gke-nodepool"
  name         = "default-pool"
  project_name = "istio-prow-build"
  location     = "us-central1-f"
  cluster_name = "prow-arm"

  initial_count = 0

  disk_size_gb    = 100
  disk_type       = "pd-ssd"
  machine_type    = "n1-standard-8"
  service_account = "istio-prow-jobs@istio-prow-build.iam.gserviceaccount.com"
}

# This pool provides the actual ARM (t2a) instances for tests.
module "prow_arm_test_spot" {
  source = "../modules/gke-nodepool"

  name         = "t2a-spot"
  project_name = "istio-prow-build"
  location     = "us-central1-f"
  cluster_name = "prow-arm"

  # Currently, autoscaling is disabled due to ongoing networking issues on ARM.
  min_count     = 8
  max_count     = 8
  initial_count = 0

  disk_size_gb = 256
  disk_type    = "pd-ssd"
  labels = {
    testing = "test-pool"
  }

  arm          = true
  machine_type = "t2a-standard-16"
  # Spot instances are used as quota is capped for ARM nodes, and its cheaper.
  spot = true

  service_account = "istio-prow-jobs@istio-prow-build.iam.gserviceaccount.com"

}

# This pool provides the actual ARM (c4a) instances for tests.
module "prow_arm_test_spot_preview" {
  source = "../modules/gke-nodepool"

  name           = "c4a-16-spot"
  project_name   = "istio-prow-build"
  location       = "us-central1-f"
  node_locations = ["us-central1-a"]
  cluster_name   = "prow-arm"

  # Currently, autoscaling is disabled due to ongoing networking issues on ARM.
  min_count     = 1
  max_count     = 8
  initial_count = 0

  disk_size_gb = 256
  disk_type    = "hyperdisk-balanced"
  labels = {
    testing = "test-pool"
  }

  arm          = true
  machine_type = "c4a-standard-16"
  # Spot instances are used as quota is capped for ARM nodes, and its cheaper.
  spot = true

  service_account = "istio-prow-jobs@istio-prow-build.iam.gserviceaccount.com"

}

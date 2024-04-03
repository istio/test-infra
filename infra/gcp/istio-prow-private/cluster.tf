module "prow_cluster" {
  source       = "../modules/gke-cluster"
  cluster_name = "prow"
  project_name = local.project_id
  # ARM node pools are (currently) only available here, so ensure we use this region.
  cluster_location = "us-central1-f"
  release_channel  = "REGULAR"
}
module "prow_node_test" {
  source = "../modules/gke-nodepool"
  name   = "test"

  project_name = local.project_id
  cluster_name = module.prow_cluster.cluster.name
  location     = module.prow_cluster.cluster.location

  machine_type  = "e2-standard-16"
  initial_count = 1
  min_count     = 1
  max_count     = 15

  disk_size_gb = 256
  disk_type    = "pd-ssd"

  service_account = module.prow_cluster.cluster_node_sa.email

  labels = {
    "testing" = "test-pool"
  }
}
module "prow_node_build" {
  source = "../modules/gke-nodepool"
  name   = "build"

  project_name = local.project_id
  cluster_name = module.prow_cluster.cluster.name
  location     = module.prow_cluster.cluster.location

  machine_type  = "n1-highmem-64"
  initial_count = 0
  min_count     = 0
  max_count     = 10

  disk_size_gb = 2000
  disk_type    = "pd-ssd"

  service_account = module.prow_cluster.cluster_node_sa.email

  labels = {
    "testing" = "build-pool"
  }
}

module "prow_node_arm" {
  source = "../modules/gke-nodepool"

  name   = "arm"
  project_name = local.project_id
  cluster_name = module.prow_cluster.cluster.name
  location     = module.prow_cluster.cluster.location

  initial_count = 0
  min_count     = 0
  max_count     = 6

  disk_size_gb = 256
  disk_type    = "pd-ssd"
  labels = {
    "testing" = "test-pool"
  }

  arm  = true
  machine_type  = "t2a-standard-16"
  # GCP is only allowing non-trivial quotas for t2a nodes using spot instances, so enable spot instances.
  spot = true

  service_account = module.prow_cluster.cluster_node_sa.email
}

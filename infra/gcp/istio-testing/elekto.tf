# Elekto hosts https://elections.istio.io/. DNS is configured in istio-io project
# Elekto runs as a GCE VM
resource "google_compute_instance" "elekto_web" {
  boot_disk {
    auto_delete = true
    device_name = "elekto-web"

    initialize_params {
      image = "https://www.googleapis.com/compute/beta/projects/debian-cloud/global/images/debian-10-buster-v20210701"
      size  = 10
      type  = "pd-balanced"
    }

    mode   = "READ_WRITE"
    source = "https://www.googleapis.com/compute/v1/projects/istio-testing/zones/us-central1-a/disks/elekto-web"
  }

  confidential_instance_config {
    enable_confidential_compute = false
  }

  machine_type = "e2-small"

  name = "elekto-web"

  network_interface {
    access_config {
      nat_ip       = "34.134.184.23"
      network_tier = "PREMIUM"
    }

    network            = "https://www.googleapis.com/compute/v1/projects/istio-testing/global/networks/default"
    network_ip         = "10.128.0.39"
    subnetwork         = "https://www.googleapis.com/compute/v1/projects/istio-testing/regions/us-central1/subnetworks/default"
    subnetwork_project = "istio-testing"
  }

  project = "istio-testing"

  reservation_affinity {
    type = "ANY_RESERVATION"
  }

  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    provisioning_model  = "STANDARD"
  }

  service_account {
    email = "450874614208-compute@developer.gserviceaccount.com"
    scopes = [
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring.write",
      "https://www.googleapis.com/auth/service.management.readonly",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/trace.append",
    ]
  }

  shielded_instance_config {
    enable_integrity_monitoring = true
    enable_vtpm                 = true
  }

  tags = ["elekto-web", "http-server", "https-server"]
  zone = "us-central1-a"

  lifecycle {
    ignore_changes = [metadata["ssh-keys"]]
  }
}

# Allow various folks SSH access. Likely this was just for initial setup and we can actually remove these?
resource "google_compute_firewall" "elekto_ssh_craigbox" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }

  direction     = "INGRESS"
  name          = "elekto-ssh-craigbox"
  network       = "https://www.googleapis.com/compute/v1/projects/istio-testing/global/networks/default"
  priority      = 1000
  project       = "istio-testing"
  source_ranges = ["82.18.160.88"]
  target_tags   = ["elekto-web"]
}
resource "google_compute_firewall" "elekto_ssh_jberkus" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }

  direction     = "INGRESS"
  name          = "elekto-ssh-jberkus"
  network       = "https://www.googleapis.com/compute/v1/projects/istio-testing/global/networks/default"
  priority      = 1000
  project       = "istio-testing"
  source_ranges = ["71.237.176.63"]
  target_tags   = ["elekto-web"]
}

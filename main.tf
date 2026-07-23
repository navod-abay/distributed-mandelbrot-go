terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.8.0"
    }
  }
}

data "local_file" "github_private_key" {
  filename = "${path.module}/keys/id_ed25519"
}


variable "client_count" {
  type        = number
  description = "Number of clients"
  default     = 1 # Optional: Used if no value is passed
}

provider "google" {
  project = "terraform-learn-501309"
  region  = "asia-south2"
  zone    = "asia-south2-a"
}


data "google_client_config" "current" {

}

resource "google_service_account" "benchmark_runner" {
  account_id   = "benchmark-runner-sa"
  display_name = "Service Account"
}

resource "google_project_iam_member" "secret_accessor" {
  project = data.google_client_config.current.project
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.benchmark_runner.email}"
}

resource "google_compute_network" "vpc_network" {
  name                    = "terraform-network"
  auto-create-subnetworks = false
}

resource "google_compute_firewall" "allow_iap_ingress" {
  name    = "allow-ingress-from-iap"
  network = google_compute_network.vpc_network.name # Replace with your VPC network name

  direction     = "INGRESS"
  priority      = 1000
  source_ranges = ["35.235.240.0/20"]

  # Protocols and ports to allow through the IAP tunnel
  allow {
    protocol = "tcp"
    ports    = ["22", "3389"] # 22 for SSH, 3389 for RDP
  }
}

resource "google_compute_router" "router" {
  name    = "my-router"
  network = google_compute_network.vpc_network.name
}

resource "google_compute_router_nat" "nat" {
  name   = "my-nat"
  router = google_compute_router.router.name

  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}

resource "google_compute_subnetwork" "private_subnet" {
  name          = "subnet-1"
  region        = "asia-south2"
  network       = google_compute_network.vpc_network.id
  ip_cidr_range = "192.168.1.0/24"
}

resource "google_secret_manager_secret" "github_private_key" {
  secret_id = "github_private_key"

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "github_private_key_v1" {
  secret      = google_secret_manager_secret.github_private_key.id
  secret_data = data.local_file.github_private_key.content
}

resource "google_secret_manager_secret_iam_member" "vm_accessor" {
  secret_id = google_secret_manager_secret.github_private_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.benchmark_runner.email}"
}


resource "google_compute_instance" "master" {
  name         = "master-node"
  machine_type = "e2-standard-2"

  scheduling {
    provisioning_model          = "SPOT"
    preemptible                 = true
    automatic_restart           = false
    instance_termination_action = "STOP"
  }
  network_interface {
    network    = google_compute_network.vpc_network.name
    subnetwork = google_compute_subnetwork.private_subnet.name
  }

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
    }
  }

  service_account {
    email  = google_service_account.benchmark_runner.email
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  metadata = {
    TF_VAR_client_count    = var.client_count
    TF_VAR_vpc_name        = google_compute_network.vpc_network.name
    TF_VAR_subnetwork_name = google_compute_subnetwork.private_subnet.name
    TF_VAR_project_name    = data.google_client_config.current.project
  }

  metadata_startup_script = file("${path.module}/bootstrap_master.sh")
}

output "master_ip" {
  value       = google_compute_instance.master.network_interface[0].network_ip
  description = "IP adress of the master node"
  sensitive   = false
  depends_on  = [google_compute_instance.master]
}
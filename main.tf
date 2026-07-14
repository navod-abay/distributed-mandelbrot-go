terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.8.0"
    }
  }
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

resource "google_compute_network" "vpc_network" {
  name = "terraform-network"
}

resource "google_compute_subnetwork" "private_subnet" {
  name          = "subnet-1"
  region        = "asia-south2"
  network       = google_compute_network.vpc_network.id
  ip_cidr_range = "192.168.1.0/24"
}

resource "google_compute_instance" "client" {
  name         = "client-node"
  count = var.client_count
  machine_type = "e2-standard-16"

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
}

resource "google_compute_instance" "master" {
  name         = "master-node"
  machine_type = "e2-standard-16"

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

  metadata_startup_script =  file("${path.module}/bootstrap.sh")
}

output "client_ip" {
  value       = google_compute_instance.client[0].network_interface[0].network_ip
  description = "IP adress of the client node"
  sensitive   = false
  depends_on  = [google_compute_instance.client]
}
output "master_ip" {
  value       = google_compute_instance.master.network_interface[0].network_ip
  description = "IP adress of the master node"
  sensitive   = false
  depends_on  = [google_compute_instance.master]
}
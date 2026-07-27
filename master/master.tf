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

variable "vpc_name" {
  type        = string
  description = "Name of the VPC"
  default     = "default" # Optional: Used if no value is passed
}

variable "subnetwork_name" {
  type        = string
  description = "Name of the Subnet"
  default     = "subnet-1" # Optional: Used if no value is passed
}
provider "google" {
  project = "terraform-learn-501309"
  region  = "asia-south2"
  zone    = "asia-south2-a"
}


resource "google_compute_instance" "client" {
  name         = "client-node"
  count        = var.client_count
  machine_type = "e2-standard-2"

  scheduling {
    provisioning_model          = "SPOT"
    preemptible                 = true
    automatic_restart           = false
    instance_termination_action = "STOP"
  }
  network_interface {
    network    = var.vpc_name
    subnetwork = var.subnetwork_name
  }

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
    }
  }
}

output "client_ips" {
  value       = google_compute_instance.client[*].network_interface[0].network_ip
  description = "IP adress of the client node"
  sensitive   = false
  depends_on  = [google_compute_instance.client]
}


resource "local_file" "ansible_inventory" {
  content = templatefile("${path.module}/hosts.ini.tftpl", {
    worker_ips = {
      for instance in google_compute_instance.client :
      instance.name => instance.network_interface[0].network_ip[0]
    }
  })
  
  filename        = "${path.module}/inventory.ini"
  file_permission = "0644"
}

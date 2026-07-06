terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.8.0"
    }
  }
}

provider "google" {
  project = "terraform-learn-501309"
  region  = "asia-south2"
  zone    = "asia-south2-a"
}

resource "google_compute_network" "vpc_network" {
  name = "terraform-network"
}

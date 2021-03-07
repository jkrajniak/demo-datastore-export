terraform {
  backend "gcs" {
    credentials = "key.json"
    bucket = "datastore-tf-state-dev"
    prefix = "terraform/state"
  }
  required_providers {
    archive = {
      version = "~> 2.0.0"
      source = "hashicorp/archive"
    }
    google = {
      version = "~> 3.49.0"
      source = "hashicorp/google"
    }
  }
}

provider "google" {
  project = var.project_id
  region = var.region
}
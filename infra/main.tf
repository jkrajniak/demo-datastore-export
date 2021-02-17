terraform {
  backend "gcs" {
    credentials = "key.json"
    bucket = "datastore-tf-state-dev"
    prefix = "terraform/state"
  }
  required_version = "0.14.2"
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

resource "google_pubsub_topic" "trigger-topic" {
  name = "data-sync-trigger-topic-${var.stage}"
}

resource "google_cloud_scheduler_job" "job" {
  name = "data-sync-job-${var.stage}"
  description = "trigger data sync"
  schedule = "0 6 * * *"
  time_zone = "Europe/Warsaw"

  pubsub_target {
    topic_name = google_pubsub_topic.trigger-topic.id
    data = base64encode("{}")
  }
}

resource "google_storage_bucket" "datastore_output_bucket" {
  name = "${var.project_id}-${var.stage}-datastore-export-bucket"
  location = "EU"
  force_destroy = true
}

resource "google_bigquery_dataset" "bigquery_output_dataset" {
  dataset_id = "datastore"
  location = "EU"
}

resource "google_storage_bucket" "gcf_bucket" {
  name = "${var.project_id}-${var.stage}-gcf-bucket"
  force_destroy = true
}
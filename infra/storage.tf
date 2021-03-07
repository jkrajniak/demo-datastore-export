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
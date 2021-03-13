resource "google_service_account" "gcf_datastore_bigquery_loader" {
  account_id = "gcf-datastore-bq-loader-${var.stage}"
  display_name = "gcf-datastore-bigquery-loader-${var.stage}"
}
resource "google_project_iam_member" "gcf_datastore_bigquery" {
  for_each = toset([
    "roles/monitoring.metricWriter",
    "roles/bigquery.jobUser",
    "roles/logging.logWriter",
  ])
  role = each.value
  member = "serviceAccount:${google_service_account.gcf_datastore_bigquery_loader.email}"
}
resource "google_bigquery_dataset_iam_binding" "bigquery_output_dataset" {
  dataset_id = google_bigquery_dataset.bigquery_output_dataset.dataset_id
  role = "roles/bigquery.dataEditor"
  members = [
    "serviceAccount:${google_service_account.gcf_datastore_bigquery_loader.email}"
  ]
}
resource "google_storage_bucket_iam_binding" "datastore_output_bucket" {
  bucket = google_storage_bucket.datastore_output_bucket.name
  role = "roles/storage.legacyObjectReader"
  members = [
    "serviceAccount:${google_service_account.gcf_datastore_bigquery_loader.email}"
  ]
}

// GCF Datastore Exporter

resource "google_service_account" "gcf_datastore_exporter" {
  account_id = "gcf-datastore-exporter-${var.stage}"
  display_name = "gcf-datastore-exporter-${var.stage}"
}
resource "google_project_iam_member" "gcf_datastore_exporter" {
  for_each = toset([
    "roles/monitoring.metricWriter",
    "roles/logging.logWriter",
    "roles/datastore.importExportAdmin",
    "roles/datastore.viewer"
  ])
  role = each.value
  member = "serviceAccount:${google_service_account.gcf_datastore_exporter.email}"
}

// Scheduler SA
resource "google_service_account" "scheduler" {
  account_id = "scheduler-${var.stage}"
  display_name = "scheduler-${var.stage}"
}
resource "google_cloudfunctions_function_iam_binding" "scheduler-exporter" {
  cloud_function = google_cloudfunctions_function.gcf_datastore_exporter.name
  role = "roles/cloudfunctions.invoker"
  members = ["serviceAccount:${google_service_account.scheduler.email}"]
}
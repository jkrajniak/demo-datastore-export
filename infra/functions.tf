data "archive_file" "gcf_datastore_exporter_zip" {
  type = "zip"
  source_dir = "cloud_functions"
  output_path = "datastore-exporter-gcf.zip"
}

resource "google_storage_bucket_object" "gcf_datastore_exporter" {
  name = "datastore-exporter-gcf-${data.archive_file.gcf_datastore_exporter_zip.output_sha}.zip"
  bucket = google_storage_bucket.gcf_bucket.name
  source = data.archive_file.gcf_datastore_exporter_zip.output_path
}

resource "google_cloudfunctions_function" "gcf_datastore_exporter" {
  name = "datastore-export-fn-${var.stage}"
  runtime = "go113"
  source_archive_bucket = google_storage_bucket.gcf_bucket.name
  source_archive_object = google_storage_bucket_object.gcf_datastore_exporter.name
  entry_point = "DatastoreExport"
  service_account_email = google_service_account.gcf_datastore_exporter.email

  trigger_http = true

  environment_variables = {
    GCP_PROJECT = var.project_id
    OUTPUT_BUCKET = google_storage_bucket.datastore_output_bucket.name
  }
}

resource "google_cloudfunctions_function" "gcf_datastore_bigquery_loader" {
  name = "bigquery-importer-fn-${var.stage}"
  runtime = "go113"
  source_archive_bucket = google_storage_bucket.gcf_bucket.name
  source_archive_object = google_storage_bucket_object.gcf_datastore_exporter.name
  entry_point = "WatchBucket"
  service_account_email = google_service_account.gcf_datastore_bigquery_loader.email

  event_trigger {
    event_type = "google.storage.object.finalize"
    resource = google_storage_bucket.datastore_output_bucket.name
  }

  environment_variables = {
    GCP_PROJECT = var.project_id
    OUTPUT_BUCKET = google_storage_bucket.datastore_output_bucket.name
    OUTPUT_DATASET = google_bigquery_dataset.bigquery_output_dataset.dataset_id
  }
}
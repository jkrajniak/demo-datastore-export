data "archive_file" "gcf_datastore_exporter_zip" {
  type = "zip"
  source_dir = "cloud_functions"
  output_path = "datastore-exporter-gcf.zip"
}

resource "google_storage_bucket_object" "gcf_datastore_bigquery_loader" {
  name = "datastore-exporter-gcf-${data.archive_file.gcf_datastore_exporter_zip.output_sha}.zip"
  bucket = google_storage_bucket.gcf_bucket.name
  source = data.archive_file.gcf_datastore_exporter_zip.output_path
}

resource "google_pubsub_topic" "gcf_datastore_bigquery_loader_topic" {
  name = "gcf-datastore-bigquery-loader-topic-${var.stage}"
}

resource "google_cloudfunctions_function" "gcf_datastore_bigquery_loader" {
  name = "datastore-bigquery-loader-fn-${var.stage}"
  runtime = "go113"
  source_archive_bucket = google_storage_bucket.gcf_bucket.name
  source_archive_object = google_storage_bucket_object.gcf_datastore_bigquery_loader.name
  entry_point = "DatastoreExportPubsubHandler"
  service_account_email = google_service_account.gcf_datastore_bigquery_loader.email

  event_trigger {
    event_type = "providers/cloud.pubsub/eventTypes/topic.publish"
    resource = google_pubsub_topic.gcf_datastore_bigquery_loader_topic.id
  }

  environment_variables = {
    GCP_PROJECT = var.project_id
    OUTPUT_BUCKET = google_storage_bucket.datastore_output_bucket.name
    TOPIC_ID = google_pubsub_topic.gcf_datastore_bigquery_loader_topic.name
    OUTPUT_DATASET = google_bigquery_dataset.bigquery_output_dataset.dataset_id
  }
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
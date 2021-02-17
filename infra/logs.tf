resource "google_logging_project_sink" "datastore_export_log_sink" {
  name = "gcf-datastore-export-log-sink-${var.stage}"
  project = var.project_id
  unique_writer_identity = true
  destination = "pubsub.googleapis.com/${google_pubsub_topic.gcf_datastore_bigquery_loader_topic.id}"
  filter = "protoPayload.methodName=\"google.datastore.admin.v1beta1.DatastoreAdmin.ExportEntities\""
}

// grant permission
resource "google_project_iam_binding" "datastore_export_log_sink_permisions" {
  role = "roles/pubsub.publisher"

  members = [
    google_logging_project_sink.datastore_export_log_sink.writer_identity,
  ]
}
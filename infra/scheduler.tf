resource "google_cloud_scheduler_job" "job" {
  name = "data-sync-job-${var.stage}"
  description = "trigger data sync"
  schedule = "0 6 * * *"
  time_zone = "Europe/Warsaw"

  http_target {
    http_method = "POST"
    uri = google_cloudfunctions_function.gcf_datastore_exporter.https_trigger_url
    oidc_token {
      service_account_email = google_service_account.scheduler.email
    }
    body = base64encode("{}")
  }
}
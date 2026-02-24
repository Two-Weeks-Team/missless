output "cloud_run_url" {
  value       = google_cloud_run_v2_service.missless.uri
  description = "Cloud Run service URL"
}

output "storage_bucket_name" {
  value       = google_storage_bucket.assets.name
  description = "Cloud Storage bucket name"
}

output "service_account_email" {
  value       = google_service_account.backend.email
  description = "Backend service account email"
}

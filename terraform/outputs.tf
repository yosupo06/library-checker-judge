output "gh_provider_id" {
    value = google_iam_workload_identity_pool_provider.gh.name
}

output "db_migrator_sa_email" {
    value = google_service_account.db_migrator.email
}

output "uploader_sa_email" {
    value = google_service_account.uploader.email
}
output "uploader_sa_db_name" {
    value = google_sql_user.uploader.name
}

output "main_db_connection_name" {
    value = google_sql_database_instance.main.connection_name
}

output "public_bucket_name" {
    value = google_storage_bucket.public.name
}

output "private_bucket_name" {
    value = google_storage_bucket.private.name
}

output "storage_hmac_id" {
    value = google_storage_hmac_key.main.access_id
}


output "gh_provider_id" {
    value = google_iam_workload_identity_pool_provider.gh.name
}

output "storage_editor_sa_email" {
    value = google_service_account.storage_editor.email
}

output "db_migrator_sa_email" {
    value = google_service_account.db_migrator.email
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


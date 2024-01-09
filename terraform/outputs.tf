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
    value = google_sql_user.main[google_service_account.uploader.account_id].name
}

output "judge_sa_db_name" {
    value = google_sql_user.main[google_service_account.judge.account_id].name
}

output "api_deployer_sa_email" {
    value = google_service_account.api_deployer.email
}
output "judge_deployer_sa_email" {
    value = google_service_account.judge_deployer.email
}
output "frontend_deployer_sa_email" {
    value = google_service_account.frontend_deployer.email
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

output "api_image" {
    value = "${google_artifact_registry_repository.main.location}-docker.pkg.dev/${var.gcp_project_id}/main/api"
}

output "judge_image_family" {
    value = local.judge_image_family
}

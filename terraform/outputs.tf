output "gh_provider_id" {
    value = google_iam_workload_identity_pool_provider.gh.name
}

output "db_migrator_sa_email" {
    value = google_service_account.db_migrator.email
}

output "main_db_connection_name" {
    value = google_sql_database_instance.main_test.connection_name
}

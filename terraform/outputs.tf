output "gh_workload_identity_pool_id" {
    value = google_iam_workload_identity_pool.gh.id
}

output "db_migrator_sa_email" {
    value = google_service_account.db_migrator.email
}

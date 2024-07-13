// Cloud SQL
resource "random_password" "postgres" {
  length  = 30
  special = false
}

resource "google_secret_manager_secret" "postgres_password" {
  secret_id = "database-postgres-password"

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "postgres_password" {
  secret = google_secret_manager_secret.postgres_password.id

  secret_data = random_password.postgres.result
}

resource "google_sql_database_instance" "main" {
  name             = "main"
  region           = "asia-northeast1"
  database_version = "POSTGRES_15"
  root_password    = random_password.postgres.result

  depends_on = [google_service_networking_connection.main]
  settings {
    tier = var.env == "prod" ? "db-g1-small" : "db-f1-micro"
    ip_configuration {
      ipv4_enabled    = true
      private_network = google_compute_network.main.id
    }
    database_flags {
      name  = "cloudsql.iam_authentication"
      value = "on"
    }
    backup_configuration {
      enabled = true
    }
    insights_config {
      query_insights_enabled = true
    }
  }
  deletion_protection = false
}

resource "google_sql_database" "main" {
  name     = "librarychecker"
  instance = google_sql_database_instance.main.name
}

resource "google_sql_user" "main" {
  for_each = {
    for account in [
      google_service_account.uploader,
      google_service_account.judge,
      google_service_account.api,
    ] : account.account_id => account.email
  }
  # Note: for Postgres only, GCP requires omitting the ".gserviceaccount.com" suffix
  # from the service account email due to length limits on database usernames.
  name     = trimsuffix(each.value, ".gserviceaccount.com")
  instance = google_sql_database_instance.main.name
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"
}

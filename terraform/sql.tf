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
  settings {
    tier = "db-f1-micro"
    database_flags {
      name  = "cloudsql.iam_authentication"
      value = "on"
    }
    backup_configuration {
      enabled = true
    }
  }
  deletion_protection = false
}

resource "google_sql_database" "main" {
  name     = "librarychecker"
  instance = google_sql_database_instance.main.name
}

# Note: for Postgres only, GCP requires omitting the ".gserviceaccount.com" suffix
# from the service account email due to length limits on database usernames.
resource "google_sql_user" "uploader" {
  name     = trimsuffix(google_service_account.uploader.email, ".gserviceaccount.com")
  instance = google_sql_database_instance.main.name
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"
}
resource "google_sql_user" "judge" {
  name     = trimsuffix(google_service_account.judge.email, ".gserviceaccount.com")
  instance = google_sql_database_instance.main.name
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"
}

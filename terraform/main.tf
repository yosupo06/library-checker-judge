terraform {
  cloud {
    organization = "yosupo06-org"

    workspaces {
      tags = ["library-checker"]
    }
  }
}

locals {
  github_repo_owner = "yosupo06"
  github_repo_judge = "library-checker-judge"
}

provider "google" {
  project = var.gcp_project_id
  region  = "global"
}

// Cloud Storage
resource "google_storage_bucket" "public" {
  name                        = "v2-${var.env}-library-checker-data-public"
  location                    = "asia-northeast1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = "true"
}
resource "google_storage_bucket_iam_member" "public" {
  bucket = google_storage_bucket.public.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}
resource "google_storage_bucket" "private" {
  name                        = "v2-${var.env}-library-checker-data-private"
  location                    = "asia-northeast1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = "true"

  public_access_prevention = "enforced"
}

// Workload Identity
resource "google_iam_workload_identity_pool" "gh" {
  workload_identity_pool_id = "gh-pool"
  description               = "Workload Identity Pool for Github Actions"
}
resource "google_iam_workload_identity_pool_provider" "gh" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.gh.workload_identity_pool_id
  workload_identity_pool_provider_id = "gh-provider-id"
  attribute_mapping = {
    "google.subject" = "assertion.sub",
    "attribute.repository" = "assertion.repository",
    "attribute.repository_owner" = "assertion.repository_owner",
  }
  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
  attribute_condition = "assertion.repository_owner == \"${local.github_repo_owner}\""
}

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

resource "google_sql_database_instance" "main_test" {
  name             = "main-test"
  region           = "asia-northeast1"
  database_version = "POSTGRES_15"
  root_password = random_password.postgres.result
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
  name     = "library_checker"
  instance = google_sql_database_instance.main_test.name
}


resource "google_service_account" "db_owner" {
  account_id   = "db-owner-sa"
  display_name = "Service Account for DB owner"
}


resource "google_project_iam_member" "db_owner_sa_role" {
  for_each = toset([
    "roles/cloudsql.instanceUser",
  ])
  project = var.gcp_project_id
  role = each.key
  member  = "serviceAccount:${google_service_account.db_owner.email}"
}

resource "google_sql_user" "iam_service_account_user" {
  # Note: for Postgres only, GCP requires omitting the ".gserviceaccount.com" suffix
  # from the service account email due to length limits on database usernames.
  name     = trimsuffix(google_service_account.db_owner.email, ".gserviceaccount.com")
  instance = google_sql_database_instance.main_test.name
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"
}


resource "google_service_account" "db_migrator" {
  account_id   = "db-migrator"
  display_name = "DB migrator"
}
resource "google_service_account_iam_member" "db_migrator" {
  service_account_id = google_service_account.db_migrator.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principal://iam.googleapis.com/${google_iam_workload_identity_pool.gh.name}/subject/${local.github_repo_owner}/${local.github_repo_judge}"
}
resource "google_project_iam_member" "db_migrator_sa_role" {
  for_each = toset([
    "roles/secretmanager.viewer",
    "roles/cloudsql.instanceUser",
  ])
  project = var.gcp_project_id
  role    = each.key
  member  = "serviceAccount:${google_service_account.db_migrator.email}"
}
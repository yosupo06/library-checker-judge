terraform {
  cloud {
    organization = "yosupo06-org"

    workspaces {
      tags = ["library-checker"]
    }
  }
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~>3.0.2"
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
data "google_client_config" "default" {}

// Workload Identity
resource "google_iam_workload_identity_pool" "gh" {
  workload_identity_pool_id = "gh-pool"
  description               = "Workload Identity Pool for Github Actions"
}
resource "google_iam_workload_identity_pool_provider" "gh" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.gh.workload_identity_pool_id
  workload_identity_pool_provider_id = "gh-provider-id"
  attribute_mapping = {
    "google.subject"             = "assertion.sub",
    "attribute.repository"       = "assertion.repository",
    "attribute.repository_owner" = "assertion.repository_owner",
  }
  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
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

resource "google_sql_user" "uploader" {
  # Note: for Postgres only, GCP requires omitting the ".gserviceaccount.com" suffix
  # from the service account email due to length limits on database usernames.
  name     = trimsuffix(google_service_account.uploader.email, ".gserviceaccount.com")
  instance = google_sql_database_instance.main.name
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"
}


resource "google_secret_manager_secret" "discord_announcement_webhook" {
  secret_id = "discord-announcement-webhook"

  replication {
    auto {}
  }
}

resource "google_artifact_registry_repository" "main" {
  location      = "asia-northeast1"
  repository_id = "main"
  description   = "docker repository"
  format        = "DOCKER"

  docker_config {
    immutable_tags = true
  }
}




resource "google_sql_user" "api" {
  # Note: for Postgres only, GCP requires omitting the ".gserviceaccount.com" suffix
  # from the service account email due to length limits on database usernames.
  name     = trimsuffix(google_service_account.api.email, ".gserviceaccount.com")
  instance = google_sql_database_instance.main.name
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"
}


resource "google_cloud_run_v2_service" "api" {
  name     = "api"
  location = "asia-northeast1"
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    scaling {
      max_instance_count = 2
    }

    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [google_sql_database_instance.main.connection_name]
      }
    }

    containers {
      image = "${google_artifact_registry_repository.main.location}-docker.pkg.dev/${var.gcp_project_id}/main/api"
      env {
        name  = "PG_HOST"
        value = "/cloudsql/${google_sql_database_instance.main.connection_name}"
      }
      env {
        name  = "PG_TABLE"
        value = "librarychecker"
      }
      env {
        name  = "PG_USER"
        value = "postgres"
      }
      env {
        name = "PG_PASS"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.postgres_password.secret_id
            version = "latest"
          }
        }
      }
      volume_mounts {
        name       = "cloudsql"
        mount_path = "/cloudsql"
      }
    }

    service_account = google_service_account.api.email
  }
}
resource "google_cloud_run_v2_service_iam_member" "member" {
  project  = google_cloud_run_v2_service.api.project
  location = google_cloud_run_v2_service.api.location
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

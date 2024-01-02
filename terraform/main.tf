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


resource "google_service_account" "db_owner" {
  account_id   = "db-owner-sa"
  display_name = "Service Account for DB owner"
}


resource "google_project_iam_member" "db_owner_sa_role" {
  for_each = toset([
    "roles/cloudsql.instanceUser",
  ])
  project = var.gcp_project_id
  role    = each.key
  member  = "serviceAccount:${google_service_account.db_owner.email}"
}

resource "google_sql_user" "iam_service_account_user" {
  # Note: for Postgres only, GCP requires omitting the ".gserviceaccount.com" suffix
  # from the service account email due to length limits on database usernames.
  name     = trimsuffix(google_service_account.db_owner.email, ".gserviceaccount.com")
  instance = google_sql_database_instance.main.name
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"
}


resource "google_service_account" "uploader" {
  account_id   = "uploader"
  display_name = "Uploader"
}
resource "google_service_account_iam_member" "uploader" {
  service_account_id = google_service_account.uploader.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.gh.name}/attribute.repository/${local.github_repo_owner}/${local.github_repo_judge}"
}
resource "google_project_iam_member" "uploader_sa_role" {
  for_each = toset([
    "roles/cloudsql.client",
    "roles/cloudsql.instanceUser",
    "roles/secretmanager.secretAccessor",
  ])
  project = var.gcp_project_id
  role    = each.key
  member  = "serviceAccount:${google_service_account.uploader.email}"
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

resource "google_service_account" "db_migrator" {
  account_id   = "db-migrator"
  display_name = "DB migrator"
}
resource "google_service_account_iam_member" "db_migrator" {
  service_account_id = google_service_account.db_migrator.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.gh.name}/attribute.repository/${local.github_repo_owner}/${local.github_repo_judge}"
}
resource "google_project_iam_member" "db_migrator_sa_role" {
  for_each = toset([
    "roles/cloudsql.client",
    "roles/secretmanager.secretAccessor",
  ])
  project = var.gcp_project_id
  role    = each.key
  member  = "serviceAccount:${google_service_account.db_migrator.email}"
}

resource "google_service_account" "storage_editor" {
  account_id   = "storage-editor"
  display_name = "Storage editor"
}
resource "google_project_iam_member" "storage_editor_sa_role" {
  for_each = toset([
    "roles/storage.objectUser",
  ])
  project = var.gcp_project_id
  role    = each.key
  member  = "serviceAccount:${google_service_account.storage_editor.email}"
}



resource "google_storage_hmac_key" "main" {
  service_account_email = google_service_account.storage_editor.email
}
resource "google_secret_manager_secret" "storage_hmac_key" {
  secret_id = "storage-hmac-key"

  replication {
    auto {}
  }
}
resource "google_secret_manager_secret_version" "storage_hmac_key" {
  secret = google_secret_manager_secret.storage_hmac_key.id

  secret_data = google_storage_hmac_key.main.secret
}

// Artifact registry
// 



// Cloud Run

resource "google_artifact_registry_repository" "main" {
  location      = "asia-northeast1"
  repository_id = "main"
  description   = "docker repository"
  format        = "DOCKER"

  docker_config {
    immutable_tags = true
  }
}


resource "google_service_account" "api" {
  account_id   = "api-sa"
  display_name = "Service Account for API"
}
resource "google_project_iam_member" "api_sa_role" {
  for_each = toset([
    "roles/cloudsql.client",
    "roles/cloudsql.instanceUser",
    "roles/secretmanager.secretAccessor",
  ])
  project = var.gcp_project_id
  role    = each.key
  member  = "serviceAccount:${google_service_account.api.email}"
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
        name  = "PG_PASS"
        value_source {
          secret_key_ref {
            secret = google_secret_manager_secret.postgres_password.secret_id
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
  project = google_cloud_run_v2_service.api.project
  location = google_cloud_run_v2_service.api.location
  name = google_cloud_run_v2_service.api.name
  role = "roles/run.invoker"
  member = "allUsers"
}

resource "google_service_account" "api_deployer" {
  account_id   = "api-deployer-sa"
  display_name = "Service Account for API deployer"
}
resource "google_service_account_iam_member" "api_deployer" {
  service_account_id = google_service_account.api_deployer.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.gh.name}/attribute.repository/${local.github_repo_owner}/${local.github_repo_judge}"
}
resource "google_project_iam_member" "api_deployer_sa_role" {
  for_each = toset([
    "roles/artifactregistry.writer",
  ])
  project = var.gcp_project_id
  role    = each.key
  member  = "serviceAccount:${google_service_account.api_deployer.email}"
}

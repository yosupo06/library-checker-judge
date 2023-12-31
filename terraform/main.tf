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
}

provider "google" {
  project = var.gcp_project_id
  region  = "global"
}

// Cloud Storage
resource "google_storage_bucket" "public_bucket" {
  name                        = "v1-${var.env}-library-checker-data-public"
  location                    = "asia-northeast1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = "true"
}
resource "google_storage_bucket_iam_member" "public_bucket_iam_member" {
  bucket = google_storage_bucket.public_bucket.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}
resource "google_storage_bucket" "private_bucket" {
  name                        = "v1-${var.env}-library-checker-data-private"
  location                    = "asia-northeast1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = "true"

  public_access_prevention = "enforced"
}

// Workload Identity
resource "google_iam_workload_identity_pool" "gh_pool" {
  workload_identity_pool_id = "my-gh-pool"
  description               = "Workload Identity Pool for Github Actions"
}
resource "google_iam_workload_identity_pool_provider" "gh_provider" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.gh_pool.workload_identity_pool_id
  workload_identity_pool_provider_id = "my-gh-provider-id"
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
resource "google_sql_database_instance" "lc_database" {
  name             = "lc-database"
  region           = "asia-northeast1"
  database_version = "POSTGRES_15"
  settings {
    tier = "db-f1-micro"
  }

  deletion_protection  = "true"
}

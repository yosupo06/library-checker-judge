resource "google_service_account" "api_deployer" {
  account_id   = "api-deployer-sa"
  display_name = "Service Account for API deployer"
}
resource "google_service_account" "metrics_deployer" {
  account_id   = "metrics-deployer-sa"
  display_name = "Service Account for metrics deployer"
}
resource "google_service_account" "judge_deployer" {
  account_id   = "judge-deployer-sa"
  display_name = "Service Account for Judge deployer"
}
resource "google_service_account" "frontend_deployer" {
  account_id   = "frontend-deployer"
  display_name = "Service Account for Frontend deployer"
}
resource "google_service_account" "uploader" {
  account_id   = "uploader"
  display_name = "Uploader"
}
resource "google_service_account" "db_migrator" {
  account_id   = "db-migrator"
  display_name = "DB migrator"
}
resource "google_service_account" "storage_editor" {
  account_id   = "storage-editor"
  display_name = "Storage editor"
}
resource "google_service_account" "api" {
  account_id   = "api-sa"
  display_name = "Service Account for API"
}
resource "google_service_account" "judge" {
  account_id   = "judge-sa"
  display_name = "Service Account for Judge"
}
resource "google_service_account" "queue_metrics" {
  account_id   = "queue-metrics"
  display_name = "Service Account for queue metrics function"
}
resource "google_service_account" "queue_metrics_invoker" {
  account_id   = "queue-metrics-invoker"
  display_name = "Service Account for queue metrics invoker"
}

locals {
  accounts = [
    {
      account = google_service_account.api_deployer
      roles = [
        "roles/artifactregistry.writer",
        "roles/run.developer",
        "roles/iam.serviceAccountUser",
      ]
    },
    {
      account = google_service_account.metrics_deployer
      roles = [
        "roles/artifactregistry.writer",
        "roles/run.developer",
        "roles/iam.serviceAccountUser",
      ]
    },
    {
      account = google_service_account.judge_deployer
      roles = [
        "roles/compute.instanceAdmin",
        "roles/compute.storageAdmin",
        "roles/iam.serviceAccountUser",
        "roles/secretmanager.secretAccessor",
      ]
    },
    {
      account = google_service_account.frontend_deployer
      roles = [
        "roles/firebaseauth.admin",
        "roles/firebasehosting.admin",
        "roles/serviceusage.apiKeysViewer",
      ]
    },
    {
      account = google_service_account.uploader
      roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/secretmanager.secretAccessor",
      ]
    },
    {
      account = google_service_account.db_migrator
      roles = [
        "roles/cloudsql.client",
        "roles/secretmanager.secretAccessor",
      ]
    },
    {
      account = google_service_account.storage_editor
      roles = [
        // TODO: use weak permission
        "roles/storage.objectAdmin"
      ]
    },
    {
      account = google_service_account.api
      roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/secretmanager.secretAccessor",
      ]
    },
    {
      account = google_service_account.judge
      roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/secretmanager.secretAccessor",
        "roles/monitoring.metricWriter",
        "roles/logging.logWriter",
      ]
    },
    {
      account = google_service_account.queue_metrics
      roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/secretmanager.secretAccessor",
        "roles/monitoring.metricWriter",
        "roles/logging.logWriter",
      ]
    },
  ]
}

resource "google_project_iam_member" "sa_role" {
  for_each = {
    for elem in flatten([
      for account in local.accounts : [
        for role in account.roles : {
          account_id : account.account.account_id,
          email = account.account.email,
          role  = role
        }
    ]]) : "${elem.account_id}.${elem.role}" => elem
  }

  project = var.gcp_project_id
  role    = each.value.role
  member  = "serviceAccount:${each.value.email}"
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


resource "google_service_account_iam_member" "workload_identity" {
  for_each = {
    for account in [
      google_service_account.api_deployer,
      google_service_account.judge_deployer,
      google_service_account.frontend_deployer,
      google_service_account.uploader,
      google_service_account.db_migrator,
    ] : account.account_id => account.name
  }
  service_account_id = each.value
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.gh.name}/attribute.repository/${local.github_repo_owner}/${local.github_repo_judge}"
}

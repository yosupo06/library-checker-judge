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

  judge_image_family = "v3-judge-image"
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

// Instance template

resource "google_compute_image" "judge_dummy" {
  name   = "v3-judge-image-0000"
  family = local.judge_image_family

  source_image = "projects/ubuntu-os-cloud/global/images/ubuntu-2204-jammy-v20231213a"
}

data "google_compute_image" "judge" {
  family      = local.judge_image_family
  most_recent = true
  depends_on  = [google_compute_image.judge_dummy]
}

resource "google_compute_instance_template" "judge" {
  name_prefix = "judge-template-"
  description = "This template is used to create judge server."
  region = "asia-northeast1"

  machine_type   = "c2-standard-4"
  can_ip_forward = false

  // Create a new boot disk from an image
  disk {
    source_image = data.google_compute_image.judge.self_link
    auto_delete  = true
    boot         = true
    disk_type    = "pd-standard"
    disk_size_gb = 50
  }

  network_interface {
    subnetwork = google_compute_subnetwork.main.name
  }

  metadata = {
    env = var.env
  }

  service_account {
    email  = google_service_account.judge.email
    scopes = ["cloud-platform"]
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_region_instance_group_manager" "judge" {
  name = "judge"

  base_instance_name = "judge"
  region             = "asia-northeast1"

  update_policy {
    type = "PROACTIVE"
    minimal_action = "REPLACE"
    max_unavailable_fixed = 3
  }
  version {
    instance_template = google_compute_instance_template.judge.self_link_unique
  }
}

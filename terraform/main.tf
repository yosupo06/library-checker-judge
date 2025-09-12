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

  judge_image_family = "v3-judge-image"
  judge_instance_type = "c2d-highcpu-8"

  api_domain = {
    "dev" : "v2.api.dev.judge.yosupo.jp",
    "prod" : "v2.api.judge.yosupo.jp",
  }
  api_rest_domain = {
    "dev" : "v3.api.dev.judge.yosupo.jp",
    "prod" : "v3.api.judge.yosupo.jp",
  }

  region = "asia-northeast1"
  zone = "asia-northeast1-b"
}

provider "google" {
  project = var.gcp_project_id
  region  = "global"
}
data "google_project" "main" {}

resource "google_secret_manager_secret" "discord_announcement_webhook" {
  secret_id = "discord-announcement-webhook"

  replication {
    auto {}
  }
}

resource "google_artifact_registry_repository" "main" {
  location      = local.region
  repository_id = "main"
  description   = "docker repository"
  format        = "DOCKER"

  docker_config {
    immutable_tags = true
  }
}

resource "google_firebase_project" "main" {
  provider = google-beta
  project  = data.google_project.main.project_id
}

resource "google_identity_platform_config" "default" {
  project = data.google_project.main.project_id
  sign_in {
    email {
        enabled = true
    }
  }
  authorized_domains = [
    "localhost",
    "judge.yosupo.jp",
    "dev.judge.yosupo.jp",
  ]
}

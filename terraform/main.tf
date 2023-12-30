terraform {
  cloud {
    organization = "yosupo06-org"

    workspaces {
      tags = ["library-checker"]
    }
  }
}

provider "google" {
  project = var.gcp_project_id
  region  = "global"
}

resource "google_storage_bucket" "public_bucket" {
  name     = "v1-${var.env}-library-checker-data-public"
  location = "asia-northeast1"
  storage_class = "STANDARD"
  uniform_bucket_level_access = "true"
}
resource "google_storage_bucket_iam_binding" "public_bucket_iam_binding" {
  bucket = google_storage_bucket.public_bucket.name
    role = "roles/storage.objectViewer"
    members = [
      "allUsers",
    ]
}

resource "google_storage_bucket" "private_bucket" {
  name     = "v1-${var.env}-library-checker-data-private"
  location = "asia-northeast1"
  storage_class = "STANDARD"
  uniform_bucket_level_access = "true"

  public_access_prevention = "enforced"
}


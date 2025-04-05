// Cloud Storage
resource "google_storage_bucket" "public" {
  name                        = "v2-${var.env}-library-checker-data-public"
  location                    = "asia-northeast1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = "true"

  cors {
    origin = ["*"]
    method = ["GET"]
    response_header = ["Content-Type", "Access-Control-Allow-Origin"]
    max_age_seconds = 3600
  }
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
resource "google_storage_bucket" "internal" {
  for_each = toset([
    local.internal_region,
    "asia-northeast1"
  ])
  name                        = "v2-${var.env}-library-checker-${each.key}-internal"
  location                    = each.key
  storage_class               = "STANDARD"
  uniform_bucket_level_access = "true"

  public_access_prevention = "enforced"
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

// Cloud Storage
resource "google_storage_bucket" "public" {
  name                        = "v2-${var.env}-library-checker-data-public"
  location                    = "asia-northeast1"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = "true"

  cors {
    origin          = ["*"]
    method          = ["GET"]
    response_header = ["Content-Type", "Access-Control-Allow-Origin"]
    max_age_seconds = 3600
  }
}
resource "google_storage_bucket_iam_member" "public" {
  bucket = google_storage_bucket.public.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

resource "google_storage_bucket" "internal" {
  for_each = toset([
    local.region,
  ])
  name                        = "v2-${var.env}-library-checker-${each.key}-internal"
  location                    = each.key
  storage_class               = "STANDARD"
  uniform_bucket_level_access = "true"

  public_access_prevention = "enforced"
}

resource "google_storage_bucket" "private" {
  name                        = "v2-${var.env}-library-checker-data-private"
  location                    = local.region
  storage_class               = "STANDARD"
  uniform_bucket_level_access = "true"

  public_access_prevention = "enforced"
}

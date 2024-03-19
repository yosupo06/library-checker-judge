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
  region      = "asia-northeast1"

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

  labels = {
    app = "judge"
  }

  network_interface {
    subnetwork = google_compute_subnetwork.main[local.internal_region].name
  }

  scheduling {
    preemptible       = true
    automatic_restart = false
  }

  metadata = {
    env = var.env
    enable-osconfig = "TRUE"
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
  for_each = toset([
    local.internal_region,
  ])

  name = "judge-${each.key}"

  base_instance_name = "judge"
  region             = each.key

  update_policy {
    type                  = "PROACTIVE"
    minimal_action        = "REPLACE"
    max_unavailable_fixed = 3
  }
  version {
    instance_template = google_compute_instance_template.judge.self_link_unique
  }
}

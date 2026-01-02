data "google_compute_image" "debian" {
  family  = "debian-12"
  project = "debian-cloud"
}

resource "google_compute_image" "judge_dummy" {
  name   = "v3-judge-image-0000"
  family = local.judge_image_family

  source_image = data.google_compute_image.debian.id
}


data "google_compute_image" "judge" {
  family      = local.judge_image_family
  most_recent = true
  depends_on  = [google_compute_image.judge_dummy]
}

resource "google_compute_instance_template" "judge" {
  name_prefix = "judge-template-"
  description = "This template is used to create judge server."
  region      = local.region

  machine_type   = local.judge_instance_type
  can_ip_forward = false

  // Create a new boot disk from an image
  disk {
    source_image = data.google_compute_image.judge.self_link
    auto_delete  = true
    boot         = true
    disk_type    = "pd-balanced"
    disk_size_gb = 50
  }

  labels = {
    app = "judge"
  }

  network_interface {
    subnetwork = google_compute_subnetwork.main[local.region].name
  }

  scheduling {
    preemptible       = true
    automatic_restart = false
  }

  metadata = {
    env             = var.env
    enable-osconfig = "TRUE"
  }

  service_account {
    email  = google_service_account.judge.email
    scopes = ["cloud-platform"]
  }

  lifecycle {
    create_before_destroy = true
  }

  advanced_machine_features {
    threads_per_core = 1
  }
}

resource "google_compute_instance_group_manager" "judge" {
  for_each = toset([
    local.zone,
  ])

  name = "judge-${each.key}"

  base_instance_name = "judge"
  zone               = each.key

  update_policy {
    type                  = "PROACTIVE"
    minimal_action        = "REPLACE"
    max_unavailable_fixed = 3
  }
  version {
    instance_template = google_compute_instance_template.judge.self_link_unique
  }
}

resource "google_compute_autoscaler" "judge" {
  for_each = google_compute_instance_group_manager.judge

  name   = "judge-${each.key}-autoscaler"
  zone   = each.key
  target = each.value.id

  autoscaling_policy {
    min_replicas    = local.judge_autoscaling[var.env].min_replicas
    max_replicas    = local.judge_autoscaling[var.env].max_replicas
    cooldown_period = local.judge_autoscaling[var.env].cooldown_period_seconds

    metric {
      name                       = "custom.googleapis.com/judge/task_queue/pending"
      filter                     = "resource.type=\"global\""
      single_instance_assignment = local.judge_autoscaling[var.env].single_instance_assignment
    }
  }
}

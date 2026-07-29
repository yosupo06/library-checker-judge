locals {
  judge_disk_usage_alert_threshold_percent = 85
}

resource "random_password" "monitoring_webhook_auth_token" {
  length  = 32
  special = false
}

resource "google_secret_manager_secret" "monitoring_webhook_auth_token" {
  secret_id = "monitoring-webhook-auth-token"

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "monitoring_webhook_auth_token" {
  secret      = google_secret_manager_secret.monitoring_webhook_auth_token.id
  secret_data = random_password.monitoring_webhook_auth_token.result
}

resource "google_cloud_run_v2_service" "monitoring_discord_webhook" {
  name                = "monitoring-discord-webhook"
  location            = local.region
  ingress             = "INGRESS_TRAFFIC_ALL"
  deletion_protection = false

  template {
    scaling {
      min_instance_count = 0
      max_instance_count = 1
    }

    containers {
      image = "us-docker.pkg.dev/cloudrun/container/hello"
      env {
        name = "DISCORD_WEBHOOK"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.discord_alert_webhook.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "WEBHOOK_AUTH_TOKEN"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.monitoring_webhook_auth_token.secret_id
            version = "latest"
          }
        }
      }
      ports {
        container_port = 8080
      }
    }

    service_account = google_service_account.monitoring_discord_webhook.email
  }

  lifecycle {
    ignore_changes = [
      template[0].containers[0].image,
    ]
  }

  depends_on = [
    google_project_iam_member.sa_role,
    google_secret_manager_secret_version.discord_alert_webhook,
    google_secret_manager_secret_version.monitoring_webhook_auth_token,
  ]
}

resource "google_cloud_run_v2_service_iam_member" "monitoring_discord_webhook_invoker" {
  project  = google_cloud_run_v2_service.monitoring_discord_webhook.project
  location = google_cloud_run_v2_service.monitoring_discord_webhook.location
  name     = google_cloud_run_v2_service.monitoring_discord_webhook.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_monitoring_notification_channel" "discord_alert" {
  display_name = "Discord alert webhook"
  type         = "webhook_tokenauth"

  labels = {
    url = "${google_cloud_run_v2_service.monitoring_discord_webhook.uri}?auth_token=${random_password.monitoring_webhook_auth_token.result}"
  }

  depends_on = [
    google_cloud_run_v2_service_iam_member.monitoring_discord_webhook_invoker,
  ]
}

resource "google_monitoring_alert_policy" "judge_disk_usage_high" {
  display_name = "Judge disk usage high"
  combiner     = "OR"
  enabled      = true

  notification_channels = concat(
    var.monitoring_notification_channels,
    [google_monitoring_notification_channel.discord_alert.name],
  )

  documentation {
    mime_type = "text/markdown"
    content   = <<-EOT
      Judge VM disk usage has stayed above ${local.judge_disk_usage_alert_threshold_percent}% for at least 5 minutes.

      This can cause judge internal errors when Docker, containerd, or testcase extraction cannot write temporary files.
    EOT
  }

  conditions {
    display_name = "Judge disk usage above ${local.judge_disk_usage_alert_threshold_percent}%"

    condition_threshold {
      filter = join(" AND ", [
        "resource.type = \"gce_instance\"",
        "metric.type = \"agent.googleapis.com/disk/percent_used\"",
        "metric.labels.state = \"used\"",
        "metadata.user_labels.app = \"judge\"",
      ])

      comparison      = "COMPARISON_GT"
      threshold_value = local.judge_disk_usage_alert_threshold_percent
      duration        = "300s"

      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_MEAN"
      }

      trigger {
        count = 1
      }
    }
  }

  alert_strategy {
    auto_close = "1800s"
  }

  user_labels = {
    app = "judge"
    env = var.env
  }
}

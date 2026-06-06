locals {
  judge_disk_usage_alert_threshold_percent = 85
}

resource "google_monitoring_alert_policy" "judge_disk_usage_high" {
  display_name = "Judge disk usage high"
  combiner     = "OR"
  enabled      = true

  notification_channels = var.monitoring_notification_channels

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

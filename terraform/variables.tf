variable "env" {
  type = string
}

variable "gcp_project_id" {
  type = string
}

variable "monitoring_notification_channels" {
  type        = list(string)
  description = "Cloud Monitoring notification channel resource names for alert policies."
  default     = []
}

variable "discord_alert_webhook_url" {
  type        = string
  description = "Discord webhook URL for Cloud Monitoring alerts."
  sensitive   = true

  validation {
    condition     = length(trimspace(var.discord_alert_webhook_url)) > 0
    error_message = "discord_alert_webhook_url must not be empty."
  }
}

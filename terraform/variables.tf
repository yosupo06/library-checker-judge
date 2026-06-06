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

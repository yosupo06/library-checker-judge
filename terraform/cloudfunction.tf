resource "google_monitoring_metric_descriptor" "judge_pending_tasks" {
  type         = "custom.googleapis.com/judge/task_queue/pending"
  display_name = "Judge Pending Tasks"
  metric_kind  = "GAUGE"
  value_type   = "DOUBLE"
  unit         = "1"
  description  = "Pending tasks waiting in the judge queue"
  project      = var.gcp_project_id
}

resource "google_cloud_run_v2_service" "queue_metrics" {
  name     = "queue-metrics"
  location = local.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    scaling {
      min_instance_count = 0
      max_instance_count = 1
    }

    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [google_sql_database_instance.main.connection_name]
      }
    }

    containers {
      image = "us-docker.pkg.dev/cloudrun/container/hello"
      env {
        name  = "METRIC_TYPE"
        value = google_monitoring_metric_descriptor.judge_pending_tasks.type
      }
      env {
        name  = "PGHOST"
        value = "/cloudsql/${google_sql_database_instance.main.connection_name}"
      }
      env {
        name  = "PGPORT"
        value = "5432"
      }
      env {
        name  = "PGDATABASE"
        value = google_sql_database.main.name
      }
      env {
        name  = "PGUSER"
        value = "postgres"
      }
      env {
        name = "PGPASSWORD"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.postgres_password.secret_id
            version = "latest"
          }
        }
      }
      volume_mounts {
        name       = "cloudsql"
        mount_path = "/cloudsql"
      }
      ports {
        container_port = 8080
      }
    }

    service_account = google_service_account.queue_metrics.email
  }

  lifecycle {
    ignore_changes = [
      template[0].containers[0].image,
    ]
  }
}

resource "google_cloud_run_v2_service_iam_member" "queue_metrics_invoker" {
  project  = google_cloud_run_v2_service.queue_metrics.project
  location = google_cloud_run_v2_service.queue_metrics.location
  name     = google_cloud_run_v2_service.queue_metrics.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.queue_metrics_invoker.email}"
}

resource "google_service_account_iam_member" "queue_metrics_invoker_token" {
  service_account_id = google_service_account.queue_metrics_invoker.name
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = "serviceAccount:service-${data.google_project.main.number}@gcp-sa-cloudscheduler.iam.gserviceaccount.com"
}

resource "google_cloud_scheduler_job" "queue_metrics" {
  name             = "queue-metrics"
  description      = "Invoke queue metrics service every minute"
  schedule         = "* * * * *"
  region           = local.region
  attempt_deadline = "60s"

  http_target {
    uri         = google_cloud_run_v2_service.queue_metrics.uri
    http_method = "POST"
    oidc_token {
      service_account_email = google_service_account.queue_metrics_invoker.email
      audience              = google_cloud_run_v2_service.queue_metrics.uri
    }
  }
}

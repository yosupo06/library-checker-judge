resource "google_compute_network" "main" {
  name                    = "main"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "main" {
  for_each = {
    "asia-east1" : "10.0.0.0/16",
    "asia-northeast1" : "10.1.0.0/16",
    "us-east1" : "10.2.0.0/16",
  }
  name                     = "main"
  ip_cidr_range            = each.value
  region                   = each.key
  role                     = "ACTIVE"
  network                  = google_compute_network.main.id
  private_ip_google_access = true
}

resource "google_compute_firewall" "allow_ssh" {
  name    = "main-allow-ssh"
  network = google_compute_network.main.name
  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
  source_ranges = ["0.0.0.0/0"]
}
resource "google_compute_firewall" "allow_internal" {
  name    = "main-allow-internal"
  network = google_compute_network.main.name
  allow {
    protocol = "icmp"
  }
  allow {
    protocol = "tcp"
    ports    = ["0-65535"]
  }
  allow {
    protocol = "udp"
    ports    = ["0-65535"]
  }
  source_ranges = ["10.128.0.0/9"]
}

resource "google_compute_global_address" "private" {
  name          = "private-ip-address"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.main.id
}
resource "google_service_networking_connection" "main" {
  network                 = google_compute_network.main.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private.name]
}

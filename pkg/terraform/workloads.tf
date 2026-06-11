# workloads.tf - Proyecto de Servicio para ML/Apps
resource "google_project" "service_app" {
  name            = "prj-s-ml-app"
  project_id      = "prj-s-ml-app-${var.random_suffix}"
  folder_id       = google_folder.workloads.name
  billing_account = var.billing_account
}

resource "google_compute_shared_vpc_service_project" "service_attach" {
  host_project    = google_project.host_project.project_id
  service_project = google_project.service_app.project_id
}

# Endpoint PSC para Vertex AI (Acceso Privado Seguro)
resource "google_compute_address" "vertex_psc_ip" {
  name         = "addr-vertex-psc"
  project      = google_project.host_project.project_id
  subnetwork   = google_compute_subnetwork.sub_workloads.id
  address_type = "INTERNAL"
  region       = "europe-west1"
}

resource "google_compute_forwarding_rule" "vertex_psc" {
  name                  = "fwd-vertex-psc"
  project               = google_project.host_project.project_id
  region                = "europe-west1"
  target                = "all-google-apis" # O el endpoint específico de Vertex AI
  network               = google_compute_network.main_vpc.id
  ip_address            = google_compute_address.vertex_psc_ip.id
  load_balancing_scheme = "" # PSC no usa LB scheme tradicional
}

# Cloud Run con Direct VPC Egress [3, 4]
resource "google_cloud_run_v2_service" "app" {
  name     = "cr-stateless-app"
  location = "europe-west1"
  project  = google_project.service_app.project_id

  template {
    containers {
      image = "gcr.io/cloudrun/hello"
    }
    vpc_access {
      network_interfaces {
        network    = google_compute_network.main_vpc.id
        subnetwork = google_compute_subnetwork.sub_workloads.name
      }
      egress = "ALL_TRAFFIC"
    }
  }
}
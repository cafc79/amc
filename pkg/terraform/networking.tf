# networking.tf - Proyecto Host y VPC
resource "google_project" "host_project" {
  name            = "prj-h-net-prod"
  project_id      = "prj-h-net-prod-${var.random_suffix}"
  folder_id       = google_folder.networking.name
  billing_account = var.billing_account
}

resource "google_compute_shared_vpc_host_project" "host" {
  project = google_project.host_project.project_id
}

resource "google_compute_network" "main_vpc" {
  name                    = "vpc-shared-main"
  project                 = google_project.host_project.project_id
  auto_create_subnetworks = false
}

# Subred para GKE y Cloud Run (Direct VPC Egress)
resource "google_compute_subnetwork" "sub_workloads" {
  name          = "sb-prod-europe-west1"
  project       = google_project.host_project.project_id
  network       = google_compute_network.main_vpc.id
  region        = "europe-west1"
  ip_cidr_range = "10.0.1.0/24"
  
  # Requerido para GKE [1]
  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "10.100.0.0/16"
  }
}

# Cloud NAT para salida segura a Internet [2]
resource "google_compute_router" "router" {
  name    = "cr-shared-nat"
  project = google_project.host_project.project_id
  region  = "europe-west1"
  network = google_compute_network.main_vpc.id
}

resource "google_compute_router_nat" "nat" {
  name                               = "nat-shared-gateway"
  project                            = google_project.host_project.project_id
  router                             = google_compute_router.router.name
  region                             = "europe-west1"
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}
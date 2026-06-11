# main.tf - Jerarquía de Carpetas
resource "google_folder" "common" {
  display_name = "fldr-common"
  parent       = "organizations/${var.org_id}"
}

resource "google_folder" "networking" {
  display_name = "fldr-networking"
  parent       = "organizations/${var.org_id}"
}

resource "google_folder" "workloads" {
  display_name = "fldr-workloads"
  parent       = "organizations/${var.org_id}"
}

# Política de Organización: Bloquear IPs Públicas en toda la jerarquía
resource "google_organization_policy" "no_external_ips" {
  org_id     = var.org_id
  constraint = "compute.vmExternalIpAccess"

  list_policy {
    deny {
      all = true
    }
  }
}
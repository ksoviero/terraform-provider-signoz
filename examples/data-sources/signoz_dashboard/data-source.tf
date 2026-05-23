# Dashboard data source requires `id` (no name lookup).
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_dashboard" "service_overview" {
  id = "00000000-0000-0000-0000-000000000000" # required (Terraform)
}

output "service_overview_title" {
  value = jsondecode(data.signoz_dashboard.service_overview.data).title # read-only
}

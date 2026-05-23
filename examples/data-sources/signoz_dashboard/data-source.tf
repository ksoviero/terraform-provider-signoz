data "signoz_dashboard" "service_overview" {
  id = "00000000-0000-0000-0000-000000000000"
}

output "service_overview_title" {
  value = jsondecode(data.signoz_dashboard.service_overview.data).title
}

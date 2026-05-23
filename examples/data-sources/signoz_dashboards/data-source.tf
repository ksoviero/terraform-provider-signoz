data "signoz_dashboards" "all" {}

output "dashboard_count" {
  value = length(data.signoz_dashboards.all.dashboards)
}

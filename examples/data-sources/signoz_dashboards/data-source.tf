# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_dashboards" "all" {} # no arguments

output "dashboard_count" {
  value = length(data.signoz_dashboards.all.dashboards) # read-only
}

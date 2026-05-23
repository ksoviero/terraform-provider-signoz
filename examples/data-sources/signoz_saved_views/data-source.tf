# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_saved_views" "logs" {
  source_page = "logs" # optional (list filter)
}

output "saved_view_count" {
  value = length(data.signoz_saved_views.logs.views) # read-only
}

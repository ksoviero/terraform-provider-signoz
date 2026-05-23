# Lookup by name and source_page. See provider docs for optional filters on the plural data source.
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_saved_view" "error_logs" {
  name        = "terraform-error-logs" # required (lookup; exactly one of id or name)
  source_page = "logs"                 # optional (filter when looking up by name)
}

output "saved_view_id" {
  value = data.signoz_saved_view.error_logs.id # read-only
}

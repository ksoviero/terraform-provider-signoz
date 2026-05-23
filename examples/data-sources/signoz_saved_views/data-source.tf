data "signoz_saved_views" "logs" {
  source_page = "logs"
}

output "saved_view_count" {
  value = length(data.signoz_saved_views.logs.views)
}

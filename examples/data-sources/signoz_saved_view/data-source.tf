data "signoz_saved_view" "error_logs" {
  name        = "terraform-error-logs"
  source_page = "logs"
}

output "saved_view_id" {
  value = data.signoz_saved_view.error_logs.id
}

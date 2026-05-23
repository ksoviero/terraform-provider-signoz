# Saved view example for the logs explorer (`/api/v1/explorer/views`). `composite_query` must pass SigNoz validation.

resource "signoz_saved_view" "error_logs" {
  name        = "terraform-error-logs"
  source_page = "logs"
  category    = "general"
  tags        = ["terraform", "errors"]

  composite_query = jsonencode({
    queryType = "builder"
    panelType = "list"
    queries = [{
      type = "builder_query"
      spec = {
        name   = "A"
        signal = "logs"
        filter = { expression = "severity_text = 'ERROR'" }
        order  = [{ key = "timestamp", direction = "desc" }]
        limit  = 100
        offset = 0
      }
    }]
  })
}

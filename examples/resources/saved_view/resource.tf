# Saved view example for the logs explorer (`/api/v1/explorer/views`). `composite_query` must pass SigNoz validation.
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

resource "signoz_saved_view" "error_logs" {
  name        = "terraform-error-logs"  # required (Terraform)
  source_page = "logs"                  # required (Terraform)
  category    = "general"               # optional
  tags        = ["terraform", "errors"] # optional

  composite_query = jsonencode({                             # required (Terraform)
    queryType = "builder"                                    # required (SigNoz API)
    panelType = "list"                                       # required (SigNoz API)
    queries = [{                                             # required (SigNoz API)
      type = "builder_query"                                 # required
      spec = {                                               # required
        name   = "A"                                         # required
        signal = "logs"                                      # required
        filter = { expression = "severity_text = 'ERROR'" }  # optional
        order  = [{ key = "timestamp", direction = "desc" }] # optional
        limit  = 100                                         # optional
        offset = 0                                           # optional
      }
    }]
  })
}

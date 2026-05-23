# Dashboard examples (`PostableDashboard`). See docs/DASHBOARD_API.md for the HTTP API.
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

resource "signoz_dashboard" "service_overview" {
  data = jsonencode({                                                         # required (Terraform)
    title       = "Terraform service overview"                                # optional (dashboard JSON)
    description = "Request rate for a sample metric using the query builder." # optional
    version     = "v5"                                                        # optional (recommended v5)
    tags        = ["terraform", "metrics"]                                    # optional
    layout = [{                                                               # optional
      i = "a1111111-1111-4111-8111-111111111111"                              # required (widget id reference)
      x = 0                                                                   # optional
      y = 0                                                                   # optional
      w = 12                                                                  # optional
      h = 6                                                                   # optional
    }]
    widgets = [{                                           # optional
      id          = "a1111111-1111-4111-8111-111111111111" # required (matches layout.i)
      title       = "Request rate"                         # optional
      description = "sum_rate on signoz_calls_total"       # optional
      panelTypes  = "graph"                                # optional
      query = {                                            # required (widget query)
        queryType = "builder"                              # required
        builder = {                                        # required for builder panels
          queryData = [{                                   # required
            queryName         = "A"                        # required
            dataSource        = "metrics"                  # required
            aggregateOperator = "sum_rate"                 # required
            aggregateAttribute = {                         # required
              key      = "signoz_calls_total"              # required
              dataType = "float64"                         # optional
              type     = "Sum"                             # optional
              isColumn = true                              # optional
            }
            timeAggregation  = "rate"                     # optional
            spaceAggregation = "sum"                      # optional
            stepInterval     = 60                         # optional
            filters          = { items = [], op = "AND" } # optional
            groupBy          = []                         # optional
            legend           = "rps"                      # optional
            expression       = "A"                        # optional
          }]
        }
      }
      timePreferance = "GLOBAL_TIME" # optional
      yAxisUnit      = "none"        # optional
    }]
    variables = {} # optional
  })
}

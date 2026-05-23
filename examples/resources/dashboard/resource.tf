# Dashboard examples (`PostableDashboard`). See docs/DASHBOARD_API.md for the HTTP API.
# `data` is the full dashboard document: title, version, layout, widgets, tags, variables.

resource "signoz_dashboard" "service_overview" {
  data = jsonencode({
    title       = "Terraform service overview"
    description = "Request rate for a sample metric using the query builder."
    version     = "v5"
    tags        = ["terraform", "metrics"]
    layout = [{
      i = "a1111111-1111-4111-8111-111111111111"
      x = 0
      y = 0
      w = 12
      h = 6
    }]
    widgets = [{
      id          = "a1111111-1111-4111-8111-111111111111"
      title       = "Request rate"
      description = "sum_rate on signoz_calls_total"
      panelTypes  = "graph"
      query = {
        queryType = "builder"
        builder = {
          queryData = [{
            queryName         = "A"
            dataSource        = "metrics"
            aggregateOperator = "sum_rate"
            aggregateAttribute = {
              key      = "signoz_calls_total"
              dataType = "float64"
              type     = "Sum"
              isColumn = true
            }
            timeAggregation  = "rate"
            spaceAggregation = "sum"
            stepInterval     = 60
            filters          = { items = [], op = "AND" }
            groupBy          = []
            legend           = "rps"
            expression       = "A"
          }]
        }
      }
      timePreferance = "GLOBAL_TIME"
      yAxisUnit      = "none"
    }]
    variables = {}
  })
}

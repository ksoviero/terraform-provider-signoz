resource "signoz_notification_channel" "slack" {
  name = "terraform-slack"

  config = jsonencode({
    slack_configs = [{
      api_url = "https://hooks.slack.com/services/XXX/YYY/ZZZ"
      channel = "#alerts"
    }]
  })
}

resource "signoz_alert_rule" "example" {
  alert      = "Example metric alert"
  alert_type = "METRIC_BASED_ALERT"
  rule_type  = "threshold_rule"

  spec = jsonencode({
    schemaVersion = "v2alpha1"
    version       = "v5"
    condition = {
      compositeQuery = {
        queryType = "builder"
        panelType = "graph"
        queries = [{
          type = "builder_query"
          spec = {
            name         = "A"
            signal       = "metrics"
            stepInterval = 60
            aggregations = [{ expression = "count()" }]
            filter       = { expression = "service.name = 'api'" }
          }
        }]
      }
      selectedQueryName = "A"
      thresholds = {
        kind = "basic"
        spec = [{
          name      = "critical"
          target    = 0
          op        = "above"
          matchType = "at_least_once"
          channels  = [signoz_notification_channel.slack.name]
        }]
      }
    }
    evaluation = {
      kind = "rolling"
      spec = {
        evalWindow = "5m"
        frequency  = "1m"
      }
    }
  })

  depends_on = [signoz_notification_channel.slack]
}

# Alert rule examples by signal type. Each `spec` uses schemaVersion v2alpha1 and version v5.
# Threshold operators and match types: see docs/resources/alert_rule.md (`spec` attribute) and SigNoz OpenAPI rule types.

resource "signoz_notification_channel" "slack" {
  name = "terraform-slack"

  config = jsonencode({
    slack_configs = [{
      api_url = "https://hooks.slack.com/services/XXX/YYY/ZZZ"
      channel = "#alerts"
    }]
  })
}

resource "signoz_alert_rule" "metric_cpu" {
  alert       = "Pod CPU above 80% of request"
  alert_type  = "METRIC_BASED_ALERT"
  rule_type   = "threshold_rule"
  description = "CPU usage for api-service pods exceeds 80% of the requested CPU."

  labels = {
    severity = "warning"
    team     = "platform"
  }

  annotations = {
    summary     = "Pod CPU above 80% of request"
    description = "Pod {{$labels.k8s.pod.name}} CPU is high in {{$labels.deployment.environment}}."
  }

  spec = jsonencode({
    schemaVersion = "v2alpha1"
    version       = "v5"
    condition = {
      compositeQuery = {
        queryType = "builder"
        panelType = "graph"
        unit      = "percentunit"
        queries = [{
          type = "builder_query"
          spec = {
            name         = "A"
            signal       = "metrics"
            stepInterval = 60
            aggregations = [{
              metricName       = "k8s.pod.cpu_request_utilization"
              timeAggregation  = "avg"
              spaceAggregation = "max"
            }]
            filter = { expression = "k8s.deployment.name = 'api-service'" }
            groupBy = [{
              name          = "k8s.pod.name"
              fieldContext  = "resource"
              fieldDataType = "string"
            }]
          }
        }]
      }
      selectedQueryName = "A"
      thresholds = {
        kind = "basic"
        spec = [{
          name      = "critical"
          op        = "above"
          matchType = "all_the_times"
          target    = 0.8
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
    notificationSettings = {
      renotify = {
        enabled  = false
        interval = "30m"
      }
      groupBy = ["k8s.pod.name"]
    }
  })

  depends_on = [signoz_notification_channel.slack]
}

resource "signoz_alert_rule" "logs_panic" {
  alert       = "Payments service panic logs"
  alert_type  = "LOGS_BASED_ALERT"
  rule_type   = "threshold_rule"
  description = "Any panic log line emitted by the payments service."

  labels = {
    severity = "critical"
    team     = "payments"
  }

  annotations = {
    summary     = "Payments service panic"
    description = "Panic logs detected for service payments-api."
  }

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
            signal       = "logs"
            stepInterval = 60
            aggregations = [{ expression = "count()" }]
            filter = {
              expression = "service.name = 'payments-api' AND severity_text = 'ERROR' AND body CONTAINS 'panic'"
            }
            groupBy = [{
              name          = "k8s.pod.name"
              fieldContext  = "resource"
              fieldDataType = "string"
            }]
          }
        }]
      }
      selectedQueryName = "A"
      thresholds = {
        kind = "basic"
        spec = [{
          name      = "critical"
          op        = "above"
          matchType = "at_least_once"
          target    = 0
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
    notificationSettings = {
      renotify = {
        enabled  = false
        interval = "30m"
      }
      groupBy = ["k8s.pod.name"]
    }
  })

  depends_on = [signoz_notification_channel.slack]
}

resource "signoz_alert_rule" "traces_latency" {
  alert       = "Search API p99 latency above 5s"
  alert_type  = "TRACES_BASED_ALERT"
  rule_type   = "threshold_rule"
  description = "p99 duration of the search endpoint exceeds 5 seconds."

  labels = {
    severity = "warning"
    team     = "search"
  }

  annotations = {
    summary     = "Search-api latency degraded"
    description = "p99 latency for search-api on GET /api/v1/search crossed the threshold."
  }

  spec = jsonencode({
    schemaVersion = "v2alpha1"
    version       = "v5"
    condition = {
      compositeQuery = {
        queryType = "builder"
        panelType = "graph"
        unit      = "ns"
        queries = [{
          type = "builder_query"
          spec = {
            name         = "A"
            signal       = "traces"
            stepInterval = 60
            aggregations = [{ expression = "p99(duration_nano)" }]
            filter = {
              expression = "service.name = 'search-api' AND name = 'GET /api/v1/search'"
            }
            groupBy = [
              {
                name          = "service.name"
                fieldContext  = "resource"
                fieldDataType = "string"
              },
              {
                name          = "http.route"
                fieldContext  = "attribute"
                fieldDataType = "string"
              },
            ]
          }
        }]
      }
      selectedQueryName = "A"
      thresholds = {
        kind = "basic"
        spec = [{
          name       = "warning"
          op         = "above"
          matchType  = "at_least_once"
          target     = 5
          targetUnit = "s"
          channels   = [signoz_notification_channel.slack.name]
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
    notificationSettings = {
      renotify = {
        enabled  = false
        interval = "30m"
      }
      groupBy = ["service.name", "http.route"]
    }
  })

  depends_on = [signoz_notification_channel.slack]
}

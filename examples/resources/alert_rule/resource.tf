# Alert rule examples by signal type. Each `spec` uses schemaVersion v2alpha1 and version v5.
# Threshold operators and match types: see docs/resources/alert_rule.md (`spec` attribute) and SigNoz OpenAPI rule types.
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

resource "signoz_notification_channel" "slack" {
  name = "terraform-slack" # required (Terraform)

  config = jsonencode({                                        # required (Terraform)
    slack_configs = [{                                         # required (exactly one *_configs array)
      api_url = "https://hooks.slack.com/services/XXX/YYY/ZZZ" # required (Alertmanager)
      channel = "#alerts"                                      # required (Alertmanager)
    }]
  })
}

resource "signoz_alert_rule" "metric_cpu" {
  alert       = "Pod CPU above 80% of request"                                     # required (Terraform)
  alert_type  = "METRIC_BASED_ALERT"                                               # required (Terraform)
  rule_type   = "threshold_rule"                                                   # required (Terraform)
  description = "CPU usage for api-service pods exceeds 80% of the requested CPU." # optional

  labels = { # optional
    severity = "warning"
    team     = "platform"
  }

  annotations = { # optional
    summary     = "Pod CPU above 80% of request"
    description = "Pod {{$labels.k8s.pod.name}} CPU is high in {{$labels.deployment.environment}}."
  }

  spec = jsonencode({                                              # required (Terraform)
    schemaVersion = "v2alpha1"                                     # required (SigNoz API)
    version       = "v5"                                           # required (SigNoz API)
    condition = {                                                  # required (SigNoz API)
      compositeQuery = {                                           # required (SigNoz API)
        queryType = "builder"                                      # required
        panelType = "graph"                                        # optional
        unit      = "percentunit"                                  # optional
        queries = [{                                               # required
          type = "builder_query"                                   # required
          spec = {                                                 # required
            name         = "A"                                     # required
            signal       = "metrics"                               # required
            stepInterval = 60                                      # optional
            aggregations = [{                                      # required
              metricName       = "k8s.pod.cpu_request_utilization" # required
              timeAggregation  = "avg"                             # optional
              spaceAggregation = "max"                             # optional
            }]
            filter = { expression = "k8s.deployment.name = 'api-service'" } # optional
            groupBy = [{                                                    # optional
              name          = "k8s.pod.name"
              fieldContext  = "resource"
              fieldDataType = "string"
            }]
          }
        }]
      }
      selectedQueryName = "A"                                  # required (SigNoz API)
      thresholds = {                                           # required (SigNoz API)
        kind = "basic"                                         # required
        spec = [{                                              # required
          name      = "critical"                               # optional
          op        = "above"                                  # required
          matchType = "all_the_times"                          # required
          target    = 0.8                                      # required
          channels  = [signoz_notification_channel.slack.name] # optional
        }]
      }
    }
    evaluation = {        # required (SigNoz API)
      kind = "rolling"    # required
      spec = {            # required
        evalWindow = "5m" # required
        frequency  = "1m" # required
      }
    }
    notificationSettings = { # optional
      renotify = {           # optional
        enabled  = false     # optional
        interval = "30m"     # optional
      }
      groupBy = ["k8s.pod.name"] # optional
    }
  })

  depends_on = [signoz_notification_channel.slack] # optional (ordering)
}

resource "signoz_alert_rule" "logs_panic" {
  alert       = "Payments service panic logs"                         # required (Terraform)
  alert_type  = "LOGS_BASED_ALERT"                                    # required (Terraform)
  rule_type   = "threshold_rule"                                      # required (Terraform)
  description = "Any panic log line emitted by the payments service." # optional

  labels = { # optional
    severity = "critical"
    team     = "payments"
  }

  annotations = { # optional
    summary     = "Payments service panic"
    description = "Panic logs detected for service payments-api."
  }

  spec = jsonencode({                                   # required (Terraform)
    schemaVersion = "v2alpha1"                          # required (SigNoz API)
    version       = "v5"                                # required (SigNoz API)
    condition = {                                       # required (SigNoz API)
      compositeQuery = {                                # required (SigNoz API)
        queryType = "builder"                           # required
        panelType = "graph"                             # optional
        queries = [{                                    # required
          type = "builder_query"                        # required
          spec = {                                      # required
            name         = "A"                          # required
            signal       = "logs"                       # required
            stepInterval = 60                           # optional
            aggregations = [{ expression = "count()" }] # required
            filter = {                                  # optional
              expression = "service.name = 'payments-api' AND severity_text = 'ERROR' AND body CONTAINS 'panic'"
            }
            groupBy = [{ # optional
              name          = "k8s.pod.name"
              fieldContext  = "resource"
              fieldDataType = "string"
            }]
          }
        }]
      }
      selectedQueryName = "A"                                  # required (SigNoz API)
      thresholds = {                                           # required (SigNoz API)
        kind = "basic"                                         # required
        spec = [{                                              # required
          name      = "critical"                               # optional
          op        = "above"                                  # required
          matchType = "at_least_once"                          # required
          target    = 0                                        # required
          channels  = [signoz_notification_channel.slack.name] # optional
        }]
      }
    }
    evaluation = {        # required (SigNoz API)
      kind = "rolling"    # required
      spec = {            # required
        evalWindow = "5m" # required
        frequency  = "1m" # required
      }
    }
    notificationSettings = { # optional
      renotify = {           # optional
        enabled  = false     # optional
        interval = "30m"     # optional
      }
      groupBy = ["k8s.pod.name"] # optional
    }
  })

  depends_on = [signoz_notification_channel.slack] # optional (ordering)
}

resource "signoz_alert_rule" "traces_latency" {
  alert       = "Search API p99 latency above 5s"                        # required (Terraform)
  alert_type  = "TRACES_BASED_ALERT"                                     # required (Terraform)
  rule_type   = "threshold_rule"                                         # required (Terraform)
  description = "p99 duration of the search endpoint exceeds 5 seconds." # optional

  labels = { # optional
    severity = "warning"
    team     = "search"
  }

  annotations = { # optional
    summary     = "Search-api latency degraded"
    description = "p99 latency for search-api on GET /api/v1/search crossed the threshold."
  }

  spec = jsonencode({                                              # required (Terraform)
    schemaVersion = "v2alpha1"                                     # required (SigNoz API)
    version       = "v5"                                           # required (SigNoz API)
    condition = {                                                  # required (SigNoz API)
      compositeQuery = {                                           # required (SigNoz API)
        queryType = "builder"                                      # required
        panelType = "graph"                                        # optional
        unit      = "ns"                                           # optional
        queries = [{                                               # required
          type = "builder_query"                                   # required
          spec = {                                                 # required
            name         = "A"                                     # required
            signal       = "traces"                                # required
            stepInterval = 60                                      # optional
            aggregations = [{ expression = "p99(duration_nano)" }] # required
            filter = {                                             # optional
              expression = "service.name = 'search-api' AND name = 'GET /api/v1/search'"
            }
            groupBy = [ # optional
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
      selectedQueryName = "A"                                   # required (SigNoz API)
      thresholds = {                                            # required (SigNoz API)
        kind = "basic"                                          # required
        spec = [{                                               # required
          name       = "warning"                                # optional
          op         = "above"                                  # required
          matchType  = "at_least_once"                          # required
          target     = 5                                        # required
          targetUnit = "s"                                      # optional
          channels   = [signoz_notification_channel.slack.name] # optional
        }]
      }
    }
    evaluation = {        # required (SigNoz API)
      kind = "rolling"    # required
      spec = {            # required
        evalWindow = "5m" # required
        frequency  = "1m" # required
      }
    }
    notificationSettings = { # optional
      renotify = {           # optional
        enabled  = false     # optional
        interval = "30m"     # optional
      }
      groupBy = ["service.name", "http.route"] # optional
    }
  })

  depends_on = [signoz_notification_channel.slack] # optional (ordering)
}

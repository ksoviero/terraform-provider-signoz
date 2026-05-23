# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_alert_rules" "all" {} # no arguments

output "alert_count" {
  value = length(data.signoz_alert_rules.all.rules) # read-only
}

# Lookup by alert title. Exactly one of `id` or `alert` is required.
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_alert_rule" "failed_pods" {
  alert = "Failed Pods" # required (lookup; exactly one of id or alert)
}

output "failed_pods_id" {
  value = data.signoz_alert_rule.failed_pods.id # read-only
}

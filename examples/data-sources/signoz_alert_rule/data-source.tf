data "signoz_alert_rule" "failed_pods" {
  alert = "Failed Pods"
}

output "failed_pods_id" {
  value = data.signoz_alert_rule.failed_pods.id
}

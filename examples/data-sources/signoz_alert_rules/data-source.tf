data "signoz_alert_rules" "all" {}

output "alert_count" {
  value = length(data.signoz_alert_rules.all.rules)
}

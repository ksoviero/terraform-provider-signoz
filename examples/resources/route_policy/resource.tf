# Route policy example: expression matcher plus notification channel wiring.
# Requires ADMIN API access. Channel names (not UUIDs) go in `channels`.

resource "signoz_notification_channel" "slack" {
  name = "terraform-slack"

  config = jsonencode({
    slack_configs = [{
      api_url = "https://hooks.slack.com/services/XXX/YYY/ZZZ"
      channel = "#alerts"
    }]
  })
}

resource "signoz_route_policy" "critical_to_slack" {
  name        = "terraform-critical-slack"
  description = "Route critical-severity alerts to Slack."
  expression  = "severity == 'critical'"
  kind        = "policy"
  channels    = [signoz_notification_channel.slack.name]
  tags        = ["terraform"]

  depends_on = [signoz_notification_channel.slack]
}

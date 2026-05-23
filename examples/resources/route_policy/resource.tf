# Route policy example: expression matcher plus notification channel wiring.
# Requires ADMIN API access. Channel names (not UUIDs) go in `channels`.
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

resource "signoz_route_policy" "critical_to_slack" {
  name        = "terraform-critical-slack"                 # required (Terraform)
  description = "Route critical-severity alerts to Slack." # optional
  expression  = "severity == 'critical'"                   # required (Terraform)
  kind        = "policy"                                   # optional
  channels    = [signoz_notification_channel.slack.name]   # required (Terraform)
  tags        = ["terraform"]                              # optional

  depends_on = [signoz_notification_channel.slack] # optional (ordering)
}

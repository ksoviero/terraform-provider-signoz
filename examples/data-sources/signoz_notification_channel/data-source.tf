# Lookup by channel name. Exactly one of `id` or `name` is required.
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_notification_channel" "slack" {
  name = "terraform-slack" # required (lookup; exactly one of id or name)
}

output "notification_channel_type" {
  value = data.signoz_notification_channel.slack.type # read-only
}

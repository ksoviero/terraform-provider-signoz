# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_notification_channels" "all" {} # no arguments

output "notification_channel_count" {
  value = length(data.signoz_notification_channels.all.channels) # read-only
}

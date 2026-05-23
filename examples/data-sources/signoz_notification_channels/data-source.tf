data "signoz_notification_channels" "all" {}

output "notification_channel_count" {
  value = length(data.signoz_notification_channels.all.channels)
}

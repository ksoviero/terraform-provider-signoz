data "signoz_notification_channel" "slack" {
  name = "terraform-slack"
}

output "notification_channel_type" {
  value = data.signoz_notification_channel.slack.type
}

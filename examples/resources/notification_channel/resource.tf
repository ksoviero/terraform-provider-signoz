resource "signoz_notification_channel" "slack" {
  name = "terraform-slack"

  config = jsonencode({
    slack_configs = [{
      api_url       = "https://hooks.slack.com/services/XXX/YYY/ZZZ"
      channel       = "#alerts"
      send_resolved = true
      title         = "{{ range .Alerts }}{{ .Annotations.summary }}\n{{ end }}"
    }]
  })
}

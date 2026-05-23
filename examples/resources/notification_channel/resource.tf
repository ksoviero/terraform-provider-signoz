# Notification channel examples (`AlertmanagertypesPostableChannel`). Each resource uses exactly one *_configs array.
# Copy the block that matches your receiver type.

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

resource "signoz_notification_channel" "email" {
  name = "terraform-email"

  config = jsonencode({
    email_configs = [{
      to            = "oncall@example.com"
      from          = "alerts@example.com"
      smarthost     = "smtp.example.com:587"
      auth_username = "alerts@example.com"
      auth_password = "your-smtp-password"
      require_tls   = true
      send_resolved = true
      headers = {
        Subject = "SigNoz: {{ .GroupLabels.alertname }}"
      }
    }]
  })
}

resource "signoz_notification_channel" "webhook" {
  name = "terraform-webhook"

  config = jsonencode({
    webhook_configs = [{
      url           = "https://hooks.example.com/signoz/alerts"
      send_resolved = true
      http_config = {
        bearer_token = "your-webhook-bearer-token"
      }
    }]
  })
}

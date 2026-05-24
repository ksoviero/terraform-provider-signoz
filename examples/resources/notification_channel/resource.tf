# Notification channel examples (`AlertmanagertypesPostableChannel` in SigNoz OpenAPI).
# Set the channel display name with the resource `name` attribute only — do not put `name` inside `config`.
# Each resource sets exactly one *_configs array. Supported keys also include:
# discord_configs, teams_configs, sns_configs, telegram_configs, pushover_configs,
# victorops_configs, wechat_configs, webex_configs, msteams_configs, msteamsv2_configs,
# jira_configs, rocketchat_configs, mattermost_configs, incidentio_configs.
# Field shapes follow Prometheus Alertmanager receiver config (snake_case in JSON).
#
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.
# OpenAPI does not mark inner receiver fields as required; channel keys use Alertmanager / SigNoz UI rules.

resource "signoz_notification_channel" "slack" {
  name = "terraform-slack" # required (Terraform)

  config = jsonencode({                                              # required (Terraform)
    slack_configs = [{                                               # required (exactly one *_configs array)
      api_url       = "https://hooks.slack.com/services/XXX/YYY/ZZZ" # required (Alertmanager)
      channel       = "#alerts"                                      # required (Alertmanager)
      send_resolved = true                                           # optional
      username      = "SigNoz"                                       # optional
      # optional
      title = "{{ range .Alerts }}{{ .Annotations.summary }}\n{{ end }}"
      # optional
      text = "{{ range .Alerts.Firing }}:fire: {{ .Annotations.description }}\n{{ end }}{{ range .Alerts.Resolved }}:white_check_mark: {{ .Annotations.description }}\n{{ end }}"
      # optional
      color  = "{{ if eq .Status \"firing\" }}danger{{ else }}good{{ end }}"
      footer = "SigNoz | signoz.example.com" # optional
    }]
  })
}

resource "signoz_notification_channel" "email" {
  name = "terraform-email" # required (Terraform)

  config = jsonencode({                    # required (Terraform)
    email_configs = [{                     # required (exactly one *_configs array)
      to            = "oncall@example.com" # required (Alertmanager; SigNoz UI)
      send_resolved = true                 # optional
      # optional: use html and/or text for the message body
      text = <<-EOT
        {{ if gt (len .Alerts.Firing) 0 }}Firing:
        {{ range .Alerts.Firing }}- {{ .Annotations.summary }}: {{ .Annotations.description }}
        {{ end }}{{ end }}
        {{ if gt (len .Alerts.Resolved) 0 }}Resolved:
        {{ range .Alerts.Resolved }}- {{ .Annotations.summary }}
        {{ end }}{{ end }}
      EOT
      # optional: SigNoz UI ships a default html template when omitted
      html = <<-EOT
        <!DOCTYPE html>
        <html>
        <body>
          <h2>{{ .Status | toUpper }}: {{ .CommonLabels.alertname }}</h2>
          {{ if gt (len .Alerts.Firing) 0 }}
          <h3>Firing ({{ .Alerts.Firing | len }})</h3>
          {{ range .Alerts.Firing }}
          <p><strong>{{ .Annotations.summary }}</strong><br/>{{ .Annotations.description }}</p>
          {{ end }}
          {{ end }}
          {{ if gt (len .Alerts.Resolved) 0 }}
          <h3>Resolved ({{ .Alerts.Resolved | len }})</h3>
          {{ range .Alerts.Resolved }}
          <p>{{ .Annotations.summary }}</p>
          {{ end }}
          {{ end }}
        </body>
        </html>
      EOT
    }]
  })
}

resource "signoz_notification_channel" "webhook" {
  name = "terraform-webhook" # required (Terraform)

  config = jsonencode({                                         # required (Terraform)
    webhook_configs = [{                                        # required (exactly one *_configs array)
      url           = "https://hooks.example.com/signoz/alerts" # required (Alertmanager)
      send_resolved = true                                      # optional
      max_alerts    = 0                                         # optional
      http_config = {                                           # optional
        bearer_token = "your-webhook-bearer-token"              # optional
      }
    }]
  })
}

resource "signoz_notification_channel" "webhook_basic_auth" {
  name = "terraform-webhook-basic-auth" # required (Terraform)

  config = jsonencode({                                         # required (Terraform)
    webhook_configs = [{                                        # required (exactly one *_configs array)
      url           = "https://hooks.example.com/signoz/alerts" # required (Alertmanager)
      send_resolved = true                                      # optional
      http_config = {                                           # optional
        basic_auth = {                                          # optional (auth method)
          username = "signoz"                                   # required when basic_auth is set
          password = "your-webhook-password"                    # required when basic_auth is set
        }
      }
    }]
  })
}

resource "signoz_notification_channel" "pagerduty" {
  name = "terraform-pagerduty" # required (Terraform)

  config = jsonencode({                                          # required (Terraform)
    pagerduty_configs = [{                                       # required (exactly one *_configs array)
      send_resolved = true                                       # optional
      routing_key   = "your-pagerduty-events-api-v2-routing-key" # required (SigNoz UI; Events API v2)
      client        = "SigNoz"                                   # optional
      client_url    = "https://signoz.example.com/alerts"        # optional
      # optional
      description = "[{{ .Status | toUpper }}{{ if eq .Status \"firing\" }}:{{ .Alerts.Firing | len }}{{ end }}] {{ .CommonLabels.alertname }} for {{ .CommonLabels.job }}"
      severity    = "{{ (index .Alerts 0).Labels.severity }}" # optional
      component   = "{{ .GroupLabels.service }}"              # optional
      group       = "{{ .GroupLabels.alertname }}"            # optional
      class       = "{{ .CommonLabels.severity }}"            # optional
      details = {                                             # optional
        firing       = "{{ .Alerts.Firing | toJson }}"        # optional (per-key)
        resolved     = "{{ .Alerts.Resolved | toJson }}"      # optional (per-key)
        num_firing   = "{{ .Alerts.Firing | len }}"           # optional (per-key)
        num_resolved = "{{ .Alerts.Resolved | len }}"         # optional (per-key)
      }
    }]
  })
}

resource "signoz_notification_channel" "opsgenie" {
  name = "terraform-opsgenie" # required (Terraform)

  config = jsonencode({                                      # required (Terraform)
    opsgenie_configs = [{                                    # required (exactly one *_configs array)
      api_key       = "your-opsgenie-api-key"                # required (SigNoz UI; Alertmanager)
      send_resolved = true                                   # optional
      message       = "{{ .CommonLabels.alertname }}"        # optional
      description   = "{{ .CommonAnnotations.description }}" # optional
      # optional
      priority = "{{ if eq (index .Alerts 0).Labels.severity \"critical\" }}P1{{ else if eq (index .Alerts 0).Labels.severity \"warning\" }}P2{{ else }}P3{{ end }}"
      source   = "SigNoz"                                     # optional
      entity   = "{{ .GroupLabels.service }}"                 # optional
      tags     = "signoz,terraform"                           # optional
      details = {                                             # optional
        alertname = "{{ .CommonLabels.alertname }}"           # optional (per-key)
        severity  = "{{ (index .Alerts 0).Labels.severity }}" # optional (per-key)
      }
    }]
  })
}

# Downtime schedule examples (`AlertmanagertypesSchedule`). Copy one resource block.
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

resource "signoz_downtime_schedule" "weekend_maintenance" {
  name        = "terraform-weekend-maintenance"                                              # required (Terraform)
  description = "Suppress notifications during a planned maintenance window for all alerts." # optional

  schedule = jsonencode({              # required (Terraform)
    timezone  = "UTC"                  # required (SigNoz API)
    startTime = "2026-06-07T06:00:00Z" # required (one-off window)
    endTime   = "2026-06-07T10:00:00Z" # required (one-off window)
  })
}

resource "signoz_downtime_schedule" "cpu_alert_maintenance" {
  name        = "terraform-cpu-alert-maintenance"                                                                             # required (Terraform)
  description = "Maintenance window scoped to one alert rule. Replace the placeholder UUID in alert_ids with a real rule id." # optional

  schedule = jsonencode({              # required (Terraform)
    timezone  = "UTC"                  # required (SigNoz API)
    startTime = "2026-06-08T06:00:00Z" # required (one-off window)
    endTime   = "2026-06-08T10:00:00Z" # required (one-off window)
  })

  alert_ids = ["00000000-0000-0000-0000-000000000001"] # optional
}

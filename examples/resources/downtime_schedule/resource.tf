# Downtime schedule examples (`AlertmanagertypesSchedule`). Copy one resource block.
# Omit `alert_ids` to apply per SigNoz API defaults; set `alert_ids` to target specific alert rule UUIDs.

resource "signoz_downtime_schedule" "weekend_maintenance" {
  name        = "terraform-weekend-maintenance"
  description = "Suppress notifications during a planned maintenance window for all alerts."

  schedule = jsonencode({
    timezone  = "UTC"
    startTime = "2026-06-07T06:00:00Z"
    endTime   = "2026-06-07T10:00:00Z"
  })
}

resource "signoz_downtime_schedule" "cpu_alert_maintenance" {
  name        = "terraform-cpu-alert-maintenance"
  description = "Maintenance window scoped to one alert rule. Replace the placeholder UUID in alert_ids with a real rule id."

  schedule = jsonencode({
    timezone  = "UTC"
    startTime = "2026-06-08T06:00:00Z"
    endTime   = "2026-06-08T10:00:00Z"
  })

  alert_ids = ["00000000-0000-0000-0000-000000000001"]
}

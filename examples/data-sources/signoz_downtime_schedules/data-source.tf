# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_downtime_schedules" "all" {} # no arguments

output "downtime_schedule_count" {
  value = length(data.signoz_downtime_schedules.all.schedules) # read-only
}

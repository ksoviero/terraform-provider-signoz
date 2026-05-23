# Lookup by schedule name. Exactly one of `id` or `name` is required.
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_downtime_schedule" "weekend" {
  name = "terraform-weekend-maintenance" # required (lookup; exactly one of id or name)
}

output "downtime_schedule_id" {
  value = data.signoz_downtime_schedule.weekend.id # read-only
}

data "signoz_downtime_schedules" "all" {}

output "downtime_schedule_count" {
  value = length(data.signoz_downtime_schedules.all.schedules)
}

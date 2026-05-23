data "signoz_downtime_schedule" "weekend" {
  name = "terraform-weekend-maintenance"
}

output "downtime_schedule_id" {
  value = data.signoz_downtime_schedule.weekend.id
}

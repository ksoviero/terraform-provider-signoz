data "signoz_route_policy" "critical" {
  name = "terraform-critical-slack"
}

output "route_policy_id" {
  value = data.signoz_route_policy.critical.id
}

data "signoz_route_policies" "all" {}

output "route_policy_count" {
  value = length(data.signoz_route_policies.all.policies)
}

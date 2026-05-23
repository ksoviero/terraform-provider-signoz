# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_route_policies" "all" {} # no arguments

output "route_policy_count" {
  value = length(data.signoz_route_policies.all.policies) # read-only
}

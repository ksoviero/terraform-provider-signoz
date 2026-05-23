# Lookup by policy name. Exactly one of `id` or `name` is required.
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_route_policy" "critical" {
  name = "terraform-critical-slack" # required (lookup; exactly one of id or name)
}

output "route_policy_id" {
  value = data.signoz_route_policy.critical.id # read-only
}

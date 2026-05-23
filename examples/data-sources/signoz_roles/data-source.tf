# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_roles" "all" {} # no arguments

output "role_count" {
  value = length(data.signoz_roles.all.roles) # read-only
}

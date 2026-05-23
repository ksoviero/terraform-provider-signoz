# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_auth_domains" "all" {} # no arguments

output "auth_domain_count" {
  value = length(data.signoz_auth_domains.all.domains) # read-only
}

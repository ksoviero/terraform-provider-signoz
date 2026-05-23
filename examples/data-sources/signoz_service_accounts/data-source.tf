# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_service_accounts" "all" {} # no arguments

output "service_account_count" {
  value = length(data.signoz_service_accounts.all.accounts) # read-only
}

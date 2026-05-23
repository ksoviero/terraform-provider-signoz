# Lookup by service account name. Exactly one of `id` or `name` is required.
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_service_account" "ci" {
  name = "terraform-ci" # required (lookup; exactly one of id or name)
}

output "service_account_id" {
  value = data.signoz_service_account.ci.id # read-only
}

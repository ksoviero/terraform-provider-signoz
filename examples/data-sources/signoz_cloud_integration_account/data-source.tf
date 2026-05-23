# Requires cloud_provider and account id.
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_cloud_integration_account" "aws" {
  cloud_provider = "aws"                                  # required (Terraform)
  id             = "00000000-0000-0000-0000-000000000001" # required (Terraform)
}

output "cloud_account_provider_id" {
  value = data.signoz_cloud_integration_account.aws.provider_account_id # read-only
}

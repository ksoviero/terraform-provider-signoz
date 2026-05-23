# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_cloud_integration_accounts" "aws" {
  cloud_provider = "aws" # required (Terraform)
}

output "aws_account_count" {
  value = length(data.signoz_cloud_integration_accounts.aws.accounts) # read-only
}

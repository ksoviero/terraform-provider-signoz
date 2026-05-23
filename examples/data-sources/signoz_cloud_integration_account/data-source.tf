data "signoz_cloud_integration_account" "aws" {
  cloud_provider = "aws"
  id             = "00000000-0000-0000-0000-000000000000" # replace with account UUID
}

output "cloud_account_provider_id" {
  value = data.signoz_cloud_integration_account.aws.provider_account_id
}

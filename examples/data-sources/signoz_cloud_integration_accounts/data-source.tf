data "signoz_cloud_integration_accounts" "aws" {
  cloud_provider = "aws"
}

output "aws_account_count" {
  value = length(data.signoz_cloud_integration_accounts.aws.accounts)
}

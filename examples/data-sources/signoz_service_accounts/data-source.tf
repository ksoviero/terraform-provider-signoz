data "signoz_service_accounts" "all" {}

output "service_account_count" {
  value = length(data.signoz_service_accounts.all.accounts)
}

data "signoz_service_account" "ci" {
  name = "terraform-ci"
}

output "service_account_id" {
  value = data.signoz_service_account.ci.id
}

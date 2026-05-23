data "signoz_roles" "all" {}

output "role_count" {
  value = length(data.signoz_roles.all.roles)
}

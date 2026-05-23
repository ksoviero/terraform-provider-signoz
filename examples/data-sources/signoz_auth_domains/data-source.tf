data "signoz_auth_domains" "all" {}

output "auth_domain_count" {
  value = length(data.signoz_auth_domains.all.domains)
}

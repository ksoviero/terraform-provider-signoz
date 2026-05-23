data "signoz_auth_domain" "saml" {
  name = "saml.example.com"
}

output "auth_domain_id" {
  value = data.signoz_auth_domain.saml.id
}

# Lookup by domain name. Exactly one of `id` or `name` is required.
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_auth_domain" "saml" {
  name = "saml.example.com" # required (lookup; exactly one of id or name)
}

output "auth_domain_id" {
  value = data.signoz_auth_domain.saml.id # read-only
}

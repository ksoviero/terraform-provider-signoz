# Lookup by role name. Exactly one of `id` or `name` is required.
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

data "signoz_role" "viewer_plus" {
  name = "terraform-viewer-plus" # required (lookup; exactly one of id or name)
}

output "role_id" {
  value = data.signoz_role.viewer_plus.id # read-only
}

# Custom role. Only `description` can be updated after create; permissions are managed in the SigNoz UI.
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

resource "signoz_role" "viewer_plus" {
  name        = "terraform-viewer-plus"                     # required (Terraform)
  description = "Custom role example managed by Terraform." # optional
}

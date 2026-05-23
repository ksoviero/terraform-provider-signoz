# Custom role. Only `description` can be updated after create; permissions are managed in the SigNoz UI.

resource "signoz_role" "viewer_plus" {
  name        = "terraform-viewer-plus"
  description = "Custom role example managed by Terraform."
}

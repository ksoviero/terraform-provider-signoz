data "signoz_role" "viewer_plus" {
  name = "terraform-viewer-plus"
}

output "role_id" {
  value = data.signoz_role.viewer_plus.id
}

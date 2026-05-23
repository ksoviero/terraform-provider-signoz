# Service account for API or automation access. The API returns a token only at create time (not in Terraform state).

resource "signoz_service_account" "ci" {
  name = "terraform-ci"
}

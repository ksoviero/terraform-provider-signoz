# Service account for API or automation access. The API returns a token only at create time (not in Terraform state).
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

resource "signoz_service_account" "ci" {
  name = "terraform-ci" # required (Terraform)
}

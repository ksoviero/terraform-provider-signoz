resource "signoz_dashboard" "example" {
  data = jsonencode({
    title   = "Terraform example"
    version = "v5"
    layout  = []
    widgets = []
  })
}

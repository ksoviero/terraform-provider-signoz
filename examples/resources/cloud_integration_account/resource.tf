resource "signoz_cloud_integration_account" "aws" {
  cloud_provider = "aws"

  config = jsonencode({
    aws = {
      deploymentRegion = "us-east-1"
      regions          = ["us-east-1"]
    }
  })

  credentials = jsonencode({
    sigNozApiUrl = "https://signoz.example.com"
    sigNozApiKey = "your-signoz-api-key"
    ingestionUrl = "https://ingest.signoz.example.com"
    ingestionKey = "your-ingestion-key"
  })
}

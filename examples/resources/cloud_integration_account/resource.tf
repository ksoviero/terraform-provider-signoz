# Cloud integration account (AWS). Required/optional comments follow provider schema and SigNoz OpenAPI where applicable.

resource "signoz_cloud_integration_account" "aws" {
  cloud_provider = "aws" # required (Terraform)

  config = jsonencode({                # required (Terraform)
    aws = {                            # required (SigNoz API; provider for this account)
      deploymentRegion = "us-east-1"   # required (SigNoz API)
      regions          = ["us-east-1"] # optional
    }
  })

  credentials = jsonencode({                           # optional (Terraform)
    sigNozApiUrl = "https://signoz.example.com"        # optional
    sigNozApiKey = "your-signoz-api-key"               # optional
    ingestionUrl = "https://ingest.signoz.example.com" # optional
    ingestionKey = "your-ingestion-key"                # optional
  })
}

# Provider bootstrap for examples. Required/optional comments follow the provider schema.

terraform {
  required_providers {
    signoz = {
      source = "ksoviero/signoz" # required (terraform block)
    }
  }
}

provider "signoz" {
  endpoint = "https://signoz.example.com" # optional (env: SIGNOZ_ENDPOINT)
  api_key  = var.signoz_api_key           # optional (env: SIGNOZ_API_KEY)
}

variable "signoz_api_key" {
  type      = string # required (variable)
  sensitive = true   # optional (variable attribute)
}

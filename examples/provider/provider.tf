terraform {
  required_providers {
    signoz = {
      source = "ksoviero/signoz"
    }
  }
}

provider "signoz" {
  endpoint = "https://signoz.example.com"
  api_key  = var.signoz_api_key
}

variable "signoz_api_key" {
  type      = string
  sensitive = true
}

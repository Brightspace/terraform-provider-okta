resource "okta_app_aws" "default" {
  name                  = "TerraformProviderTest"
  identity_provider_arn = var.identity_arn
}

variable "identity_arn" { type = string }

output "account_id" {
  value = "${okta_app_aws.default.id}"
}

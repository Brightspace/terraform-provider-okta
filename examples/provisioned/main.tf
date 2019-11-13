resource "okta_app_aws" "default" {
  name                  = "TerraformProviderTest"
  identity_provider_arn = var.identity_arn
}

resource "okta_app_aws_provision" "default" {
  application_id = okta_app_aws.default.id
  aws_access_key = var.access_key
  aws_secret_key = var.secret_key
}

variable "identity_arn" { type = string }
variable "access_key" { type = string }
variable "secret_key" { type = string }

output "account_id" {
  value = okta_app_aws.default.id
}

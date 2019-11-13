data "okta_app_saml" "default" {
  application_id = var.app_id
}

variable "app_id" { type = string }

output "saml_metadata_document" {
  value = "${data.okta_app_saml.default.saml_metadata_document}"
}

variable aws_account_id {
  default = "123456789"
}

variable aws_role {
  default = "role"
}

variable "okta_aws_user_id" {}
variable "okta_aws_user_secret" {}
variable "okta_api_key" {}
variable "okta_username" {}
variable "okta_password" {}

provider "okta" {
  okta_url       = "https://dev-750049.oktapreview.com"
  okta_admin_url = "https://dev-750049-admin.oktapreview.com"
  api_key        = "${var.okta_api_key}"
  username       = "${var.okta_username}"
  password       = "${var.okta_password}"
  org_id         = "0oaf1twj0t8kgcCs60h7"
}

resource "okta_app" "my-app" {
  name                     = "amazon_aws"
  label                    = "D2L-Terraform-Test"
  sign_on_mode             = "SAML_2_0"
  aws_environment_type     = "aws.amazon"
  group_filter             = "aws_(?${var.aws_account_id}\\d+)_(?${var.aws_role}[a-zA-Z0-9+=,.@\\-_]+)"
  login_url                = "https://console.aws.amazon.com/ec2/home"
  session_duration         = 43200
  role_value_pattern       = "arn:aws:iam::${var.aws_account_id}:saml-provider/OKTA,arn:aws:iam::${var.aws_account_id}:role/${var.aws_role}"
  identity_provider_arn    = "arn:aws:iam::852561389367:saml-provider/Okta"
  aws_okta_iam_user_id     = "${var.okta_aws_user_id}"
  aws_okta_iam_user_secret = "${var.okta_aws_user_secret}"
}

resource "okta_app_group" "my-group" {
  name        = "D2L-Terraform-Group-Test"
  description = "Test group from terraform provider"
  members     = ["sean.cowing@d2l.com", "Derek.Steinmoeller@d2l.com", "Jonathan.Beverly@D2L.com"]
  app_id      = "${okta_app.my-app.id}"
  saml_role   = "Dev-NDE-User"
}

output "saml_metadata_document" {
  value = "${okta_app.my-app.saml_metadata_document}"
}

output "okta_app_id" {
  value = "${okta_app.my-app.id}"
}

variable aws_account_id {
  default = "123456789"
}

variable aws_role {
  default = "role"
}

provider "okta" {
  okta_url       = "https://dev-750049.oktapreview.com"
  okta_admin_url = "https://dev-750049-admin.oktapreview.com"
  api_key        = "00YoJKKMsOoCs9mhrt3se1uMXW1BswwZD0UjMw9EWq"
  username       = "scowing@gmail.com"
  password       = "TheExpanseIsGreat1"
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
  aws_okta_iam_user_id     = "AKIAJNZ2MF3R75EZEMXQ"
  aws_okta_iam_user_secret = "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"
}

output "saml_metadata_document" {
  value = "${okta_app.my-app.saml_metadata_document}"
}

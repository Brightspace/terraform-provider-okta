provider "okta" {
  okta_url       = "https://dev-750049.oktapreview.com"
  okta_admin_url = "https://dev-750049-admin.oktapreview.com"
  api_key        = "${var.okta_api_key}"
  username       = "${var.okta_username}"
  password       = "${var.okta_password}"
  org_id         = "0oaf1twj0t8kgcCs60h7"
}

variable "okta_api_key" {}
variable "okta_username" {}
variable "okta_password" {}

resource "okta_user_attachment" "readonly" {
  app_id     = "0oaiquchfr6AoYiqB0h7"
  user       = "Jonathan.Beverly@d2l.com"
  role       = "ACE-Edgeville-Owner"
  saml_roles = ["ACE-Edgeville-Owner", "ACE-Edgeville-ReadOnly"]
}

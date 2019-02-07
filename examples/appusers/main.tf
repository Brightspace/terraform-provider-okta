provider "okta" {
  okta_url       = "https://dev-750049.oktapreview.com"
  okta_admin_url = "https://dev-750049-admin.oktapreview.com"
  api_key        = "${var.okta_api_key}"
  username       = "${var.okta_username}"
  password       = "${var.okta_password}"
  org_id         = "0oaf1twj0t8kgcCs60h7"
}

locals {
  users = {
    "Jonathan.Beverly@d2l.com" = "ACE-Edgeville-Owner",
    "Edvard.Kikic@d2l.com" = "ACE-Edgeville-User,ACE-Edgeville-ReadOnly",
    "sean.cowing@d2l.com" = "ACE-Edgeville-User",
    "Chris.Stavropoulos@d2l.com" = "ACE-Edgeville-User",
    "Derek.Steinmoeller@d2l.com" = "ACE-Edgeville-ReadOnly",
  }

  emails = "${keys(local.users)}"
}

variable "okta_api_key" {}
variable "okta_username" {}
variable "okta_password" {}

resource "okta_user_attachment" "readonly" {
  count      = "${length(local.emails)}"
  app_id     = "0oaiquchfr6AoYiqB0h7"
  user       = "${element(local.emails, count.index)}"
  role       = "ACE-Edgeville-Owner"
  saml_roles = "${sort(split(",", lookup(local.users, element(local.emails, count.index), "")))}"
}

variable "readonly" { type = set(string) }
variable "readonly_saml" { type = string }
variable "devs" { type = set(string) }
variable "devs_saml" { type = string }
variable "admins" { type = set(string) }
variable "admins_saml" { type = string }
variable "app_id" { type = string }

locals {
  members = setunion(var.devs, var.readonly, var.admins)
}

resource "okta_user_attachment" "users" {
  for_each = local.members

  app_id = var.app_id
  user   = each.key
  role   = var.readonly_saml
  saml_roles = toset(compact(list(
    contains(var.devs, each.key) ? var.devs_saml : "",
    contains(var.readonly, each.key) ? var.readonly_saml : "",
    contains(var.admins, each.key) ? var.admins_saml : ""
  )))
}

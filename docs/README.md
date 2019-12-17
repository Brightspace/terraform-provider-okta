# Okta Provider

The Okta provider is used to interact with Okta AWS resources.

The provider allows you to manage your Okta AWS connections. It needs to be configured with the proper credentials before it can be used.

Use the navigation below to read about the available resources.

## Example Usage

```hcl
# Configure the Okta Provider
provider "okta" {
  okta_url       = var.url
  okta_admin_url = var.admin_url
  api_key        = var.token
  username       = var.username
  password       = var.password
  org_id         = var.organization
}
```

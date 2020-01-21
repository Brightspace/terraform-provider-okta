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

# Okta Provider

The Okta provider is used to setup SSO for AWS accounts in Okta. This provider is not official, and is developed to cover a narrow slice of the Okta API. Specifically, the process of provisioning and user management of an AWS Application in Okta.

The provider allows you to manage your membership to an AWS Okta application. It needs to be configured with the proper credentials before it can be used. This provider requires both API Tokens and login credentials.

## Example Usage

```hcl
# Configure the Okta Provider
provider "okta" {}

# Create an okta AWS app
resource "okta_app_aws" "account" {
  name                  = "ACME-AwsAccount"
  identity_provider_arn = "arn:aws:iam::123412341234:saml-provider/Okta"
}
```

The okta provider is a [third party custom provider](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins). Third-party providers must be manually installed, since `terraform init` cannot automatically download them.

## Authentication

The Okta provider can be provided credentials for authentication using environment variables or static credentials.

### Static credentials

> Hard-coding credentials into any Terraform configuration is not recommended, and risks secret leakage should this file ever be committed to a public version control system

Static credentials can be provided by adding a series of in-line fields in the Okta provider block or by variables:

Usage:

```hcl
provider "okta" {
  okta_url       = "https://acme-corp.okta.com"
  okta_admin_url = "https://acme-corp-admin.okta.com"
  api_key        = "my-api-key"
  username       = "MyBotUser"
  password       = "P@ssw0rd!"
  org_id         = "7Zuii9HINkQODOW4BhRx5A/1"
}
```
### Environment variables

You can provide your credentials and URL targets via environment variables.

```hcl
provider "okta" {}
```

Usage:

```bash
export OKTA_URL = "https://acme-corp.okta.com"
export OKTA_ADMIN_URL = "https://acme-corp-admin.okta.com"
export OKTA_API_KEY = "my-api-key"
export OKTA_USERNAME = "MyBotUser"
export OKTA_PASSWORD = "P@ssw0rd!"
export OKTA_ORG_ID = "7Zuii9HINkQODOW4BhRx5A/1"
terraform plan
```

## Argument Reference

In addition to [generic `provider` arguments](https://www.terraform.io/docs/configuration/providers.html), the following arguments are supported in the Okta provider block:

- `okta_url` - (Optional) This is the Okta API BaseURL. It must be provided, but it can also be sourced from the `OKTA_URL` environment variable.
- `okta_admin_url` - (Optional) This is the Okta Admin WebUI URL. It must be provided, but it can also be sourced from the `OKTA_ADMIN_URL` environment variable.
- `api_key` - (Optional) This is the Okta API token. It must be provided, but it can also be sourced from the `OKTA_API_KEY` environment variable.
- `username` - (Optional) This is the username of a user that can log into the Admin WebUI. It must be provided, but it can also be sourced from the `OKTA_USERNAME` environment variable.
- `password` - (Optional) This is the password of a user that can log into the Admin WebUI. It must be provided, but it can also be sourced from the `OKTA_PASSWORD` environment variable.
- `org_id` - (Optional) This is the Okta ID for the organization. It must be provided, but it can also be sourced from the `OKTA_ORG_ID` environment variable.
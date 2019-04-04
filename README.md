Terraform Provider
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x
-	[Go](https://golang.org/doc/install) 1.8 (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/Brightspace/terraform-provider-okta`

```sh
$ mkdir -p $GOPATH/src/github.com/Brightspace; cd $GOPATH/src/github.com/Brightspace
$ git clone git@github.com:Brightspace/terraform-provider-okta
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/Brightspace/terraform-provider-okta
$ make build
```

Using the provider
----------------------

```terraform
provider "okta" {
  okta_url       = "https://dev-XXXXXX.oktapreview.com"
  okta_admin_url = "https://dev-XXXXXX-admin.oktapreview.com"
  api_key        = "XXXXXXXXXXXXX"
  username       = "automation-user"
  password       = "automation-password"
  org_id         = "0oaXXXXXXXXXXXX"
}

locals {
    account_id = "123412341234"
    aws_role   = "okta"
}

resource "okta_app" "my-app" {
  name                     = "amazon_aws"
  label                    = "Terraform Okta Test"
  sign_on_mode             = "SAML_2_0"
  aws_environment_type     = "aws.amazon"
  group_filter             = "aws_(?${local.account_id}\\d+)_(?${local.aws_role}[a-zA-Z0-9+=,.@\\-_]+)"
  login_url                = "https://console.aws.amazon.com/ec2/home"
  session_duration         = 43200
  role_value_pattern       = "arn:aws:iam::${local.account_id}:saml-provider/Okta,arn:aws:iam::${local.account_id}:role/${local.aws_role}"
  identity_provider_arn    = "arn:aws:iam::${local.account_id}:saml-provider/Okta"
  aws_okta_iam_user_id     = "XXXXXXXXXXXXXXXXXXXXXXXX"
  aws_okta_iam_user_secret = "XXXXXXXXXXXXXXXXXXXXXXXX"
}

```

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.8+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make bin
...
$ $GOPATH/bin/terraform-provider-okta
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

Terraform Provider
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x
-	[Go](https://golang.org/doc/install) 1.13 (to build the provider plugin)

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
To use a released provider in your Terraform environment, you will need to download the binary for your environment from the releases tab. Terraform does not offer an option for installing custom providers using terraform init.

You can see more detailed instructions for installation from the [plugins](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin) documentation. After placing it into your plugins directory, run terraform init to initialize it.

You can view the terraform examples with this provider under [examples/](examples/). If you'd like to experiment with the golang rest client for Okta, you can view the [tests/](tests/).

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine.

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

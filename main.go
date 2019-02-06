package main

import (
	"github.com/Brightspace/terraform-provider-okta/okta"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: okta.Provider})
}

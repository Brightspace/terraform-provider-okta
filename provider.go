package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"okta_url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_URL", nil),
				Description: "Okta base url",
			},
			"api_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_API_KEY", nil),
				Description: "Okta API Key",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"okta_app":           resourceApp(),
			"okta_saml_document": resourceSamlDocument(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		OktaURL: d.Get("okta_url").(string),
		APIKey:  d.Get("api_key").(string),
	}

	client := NewClient(&config)

	return client, nil
}

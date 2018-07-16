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
			"okta_admin_url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_ADMIN_URL", nil),
				Description: "Okta base admin url",
			},
			"api_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_API_KEY", nil),
				Description: "Okta API Key",
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_USERNAME", nil),
				Description: "Okta admin username",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_PASSWORD", nil),
				Description: "Okta admin password",
			},
			"org_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_ORGID", nil),
				Description: "Okta ID for organization",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"okta_app": resourceApp()},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		OktaURL:      d.Get("okta_url").(string),
		OktaAdminUrl: d.Get("okta_admin_url").(string),
		APIKey:       d.Get("api_key").(string),
		UserName:     d.Get("username").(string),
		Password:     d.Get("password").(string),
		OrgID:        d.Get("org_id").(string),
	}

	client := NewClient(&config)

	return client, nil
}

package okta

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
				Sensitive:   true,
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_USERNAME", nil),
				Description: "Okta admin username",
				Sensitive:   true,
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_PASSWORD", nil),
				Description: "Okta admin password",
				Sensitive:   true,
			},
			"org_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_ORG_ID", nil),
				Description: "Okta ID for organization",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"okta_app":             resourceApp(),
			"okta_user_attachment": resourceAppUserAttachment(),
			"okta_app_users":       resourceAppUsers(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"okta_app_saml": dataSourceAppSaml(),
		},
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
		RetryMaximum: 25,
	}

	okta, web := NewClient(&config)
	config.Okta = okta
	config.Web = web

	return config, nil
}

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
				Description: "This is the Okta API BaseURL. It must be provided, but it can also be sourced from the `OKTA_URL` environment variable.",
			},
			"okta_admin_url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_ADMIN_URL", nil),
				Description: "This is the Okta Admin WebUI URL. It must be provided, but it can also be sourced from the `OKTA_ADMIN_URL` environment variable.",
			},
			"api_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_API_KEY", nil),
				Description: "This is the Okta API token. It must be provided, but it can also be sourced from the `OKTA_API_KEY` environment variable.",
				Sensitive:   true,
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_USERNAME", nil),
				Description: "This is the username of a user that can log into the Admin WebUI. It must be provided, but it can also be sourced from the `OKTA_USERNAME` environment variable.",
				Sensitive:   true,
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_PASSWORD", nil),
				Description: "This is the password of a user that can log into the Admin WebUI. It must be provided, but it can also be sourced from the `OKTA_PASSWORD` environment variable.",
				Sensitive:   true,
			},
			"org_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_ORG_ID", nil),
				Description: "This is the Okta ID for the organization. It must be provided, but it can also be sourced from the `OKTA_ORG_ID` environment variable.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"okta_app_aws":           resourceAppAws(),
			"okta_app_aws_provision": resourceAppAwsProvision(),
			"okta_user_attachment":   resourceAppUserAttachment(),
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

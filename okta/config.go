package okta

import (
	"github.com/Brightspace/terraform-provider-okta/okta/api"
)

func NewClient(c *Config) (api.Okta, api.OktaWebClient) {
	okta := api.Okta{
		HostURL:      c.OktaURL,
		APIKey:       c.APIKey,
		RetryMaximum: c.RetryMaximum,
	}

	web := api.OktaWebClient{
		HostURL:  c.OktaURL,
		AdminURL: c.OktaAdminUrl,
		UserName: c.UserName,
		Password: c.Password,
		OrgID:    c.OrgID,
	}

	return okta, web
}

type Config struct {
	OktaURL      string
	OktaAdminUrl string
	APIKey       string
	UserName     string
	Password     string
	OrgID        string
	RetryMaximum int
	Okta         api.Okta
	Web          api.OktaWebClient
}

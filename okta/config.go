package okta

import (
	"github.com/Brightspace/terraform-provider-okta/okta/api"
	"http"
	"time"
)

func NewClient(c *Config) api.OktaClient {
	timeout := time.Duration(time.Second * 30)
	client = http.Client{
		Timeout: timeout,
	}()

	return api.OktaClient{
		OktaURL:      c.OktaURL,
		OktaAdminUrl: c.OktaAdminUrl,
		APIKey:       c.APIKey,
		UserName:     c.UserName,
		Password:     c.Password,
		OrgID:        c.OrgID,
		RetryMaximum: c.RetryMaximum,
		RestClient:   client,
	}
}

type Config struct {
	OktaURL      string
	OktaAdminUrl string
	APIKey       string
	UserName     string
	Password     string
	OrgID        string
	RetryMaximum int
}

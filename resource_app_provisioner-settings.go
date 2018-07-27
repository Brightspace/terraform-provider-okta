package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAppProvisionerSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppProvisionerSettingsCreate,
		Read:   placeholder,
		Update: placeholder,
		Delete: placeholder,

		Schema: map[string]*schema.Schema{
			"app_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"aws_okta_iam_user_id": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"aws_okta_iam_user_secret": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceAppProvisionerSettingsCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	appID := d.Get("app_id").(string)
	awsKey := d.Get("aws_okta_iam_user_id").(string)
	awsSecret := d.Get("aws_okta_iam_user_secret").(string)

	provisionErr := client.SetProvisioningSettings(appID, awsKey, awsSecret)
	if provisionErr != nil {
		return provisionErr
	}

	return nil
}

func placeholder(d *schema.ResourceData, m interface{}) error {
	return nil
}

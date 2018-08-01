package main

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAppIpdAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppIpdAttachmentCreate,
		Read:   resourceAppIpdAttachmentRead,
		Update: resourceAppIpdAttachmentCreate,
		Delete: resourceAppIpdAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"app_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"identity_provider_arn": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAppIpdAttachmentCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	appID := d.Get("app_id").(string)
	ipdArn := d.Get("identity_provider_arn").(string)

	application, err := client.ReadApplication(appID)
	if err != nil {
		return err
	}

	application.Settings.App.IdentityProviderArn = ipdArn

	_, updateErr := client.UpdateApplication(application)
	if updateErr != nil {
		return updateErr
	}

	d.SetId(appID)
	d.Set("Arn", ipdArn)

	return nil
}

func resourceAppIpdAttachmentRead(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	appID := d.Get("app_id").(string)

	readApplication, err := client.ReadApplication(appID)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", readApplication.Settings.App.IdentityProviderArn)
	return nil
}

func resourceAppIpdAttachmentDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	appID := d.Get("app_id").(string)

	application, err := client.ReadApplication(appID)
	if err != nil {
		return err
	}

	application.Settings.App.IdentityProviderArn = ""

	_, updateErr := client.UpdateApplication(application)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

package okta

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAppSaml() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAppSamlRead,

		Schema: map[string]*schema.Schema{
			"application_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique identifier of the application",
				ForceNew:    true,
			},
			"saml_metadata": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The SAML metadata",
			},
		},
	}
}

func dataSourceAppSamlRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(Config)
	client := config.Okta

	applicationID := d.Get("application_id").(string)

	log.Printf("[DEBUG] account: (AppID: %q)", applicationID)
	app, err := client.GetApplication(applicationID)
	if err != nil {
		return err
	}

	if app == nil {
		return fmt.Errorf("Could not find the application: %s", applicationID)
	}

	log.Printf("[DEBUG] saml: (AppID: %q, KeyID: %q)", app.ID, app.Credentials.Signing.KeyID)
	saml, err := client.GetSAMLMetadata(app.ID, app.Credentials.Signing.KeyID)
	if err != nil {
		return err
	}

	if saml == "" {
		return fmt.Errorf("Metadata returned is invalid: %s", applicationID)
	}

	d.SetId(app.ID)
	d.Set("saml_metadata", saml)

	return nil
}

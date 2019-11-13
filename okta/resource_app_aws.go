package okta

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAppAws() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppAwsCreate,
		Read:   resourceAppAwsRead,
		Update: resourceAppAwsUpdate,
		Delete: resourceAppAwsDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"identity_provider_arn": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"application_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"label": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"sign_on_mode": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_environment_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"login_url": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"session_duration": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_value_pattern": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"saml_metadata_document": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAppAwsCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	client := config.Okta

	name := d.Get("name").(string)
	identityArn := d.Get("identity_provider_arn").(string)

	application, err := client.CreateAwsApplication(name, identityArn)
	if err != nil {
		return err
	}

	d.SetId(application.ID)
	return resourceAppAwsRead(d, m)
}

func resourceAppAwsRead(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	client := config.Okta
	appID := d.Id()

	app, err := client.GetApplication(appID)
	if err != nil {
		return err
	}

	if app == nil {
		log.Printf("[WARN] Okta Application not found, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	saml, err := client.GetSAMLMetadata(app.ID, app.Credentials.Signing.KeyID)
	if err != nil {
		return err
	}

	d.Set("application_id", app.ID)
	d.Set("name", app.Name)
	d.Set("label", app.Label)
	d.Set("sign_on_mode", app.SignOnMode)
	d.Set("aws_environment_type", app.Settings.App.AwsEnvironmentType)
	d.Set("login_url", app.Settings.App.LoginURL)
	d.Set("identity_provider_arn", app.Settings.App.IdentityProviderArn)
	d.Set("session_duration", app.Settings.App.SessionDuration)
	d.Set("role_value_pattern", app.Settings.App.RoleValuePattern)
	d.Set("saml_metadata_document", saml)

	return nil
}

func resourceAppAwsUpdate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	client := config.Okta

	name := d.Get("name").(string)
	identityArn := d.Get("identity_provider_arn").(string)

	app, err := client.UpdateAwsApplication(d.Id(), name, identityArn)
	if err != nil {
		return err
	}

	d.SetId(app.ID)
	return resourceAppAwsRead(d, m)
}

func resourceAppAwsDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	client := config.Okta
	appID := d.Id()

	err := client.DeactivateApplication(appID)
	if err != nil {
		return err
	}

	err = client.DeleteApplication(appID)
	if err != nil {
		return err
	}

	return nil
}

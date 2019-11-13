package okta

import (
	"fmt"
	"log"
	"time"

	"github.com/Brightspace/terraform-provider-okta/okta/api"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/matryer/try"
)

const RetryWaitTimeInSeconds = 45 * time.Second

func resourceAppAwsProvision() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppAwsProvisionCreate,
		Read:   resourceAppAwsProvisionRead,
		Delete: resourceAppAwsProvisionDelete,

		Schema: map[string]*schema.Schema{
			"application_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"aws_access_key": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"aws_secret_key": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceAppAwsProvisionCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	client := config.Okta
	web := config.Web

	appId := d.Get("application_id").(string)
	awsKey := d.Get("aws_access_key").(string)
	awsSecret := d.Get("aws_secret_key").(string)

	application, err := client.GetApplication(appId)
	if err != nil {
		return err
	}

	err = try.Do(func(ampt int) (bool, error) {
		err := web.SetAWSProvisioning(application.ID, awsKey, awsSecret)
		if err != nil {
			time.Sleep(RetryWaitTimeInSeconds)
			return ampt < client.RetryMaximum, err
		}

		app, err := client.GetApplication(application.ID)
		if err != nil {
			time.Sleep(RetryWaitTimeInSeconds)
			return ampt < client.RetryMaximum, err
		}

		err = applicationIsProvisioned(app)
		if err != nil {
			time.Sleep(RetryWaitTimeInSeconds)
			return ampt < client.RetryMaximum, err
		}

		return ampt < client.RetryMaximum, nil
	})

	d.SetId(application.ID)
	return err
}

func resourceAppAwsProvisionRead(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	client := config.Okta
	appID := d.Id()

	readApplication, err := client.GetApplication(appID)
	if err != nil {
		return err
	}

	if readApplication == nil {
		log.Printf("[WARN] Okta Application %s (%q) not found, removing from state", d.Get("label").(string), d.Id())
		d.SetId("")
		return nil
	}

	samlMetadataDocument, err := client.GetSAMLMetadata(appID, readApplication.Credentials.Signing.KeyID)
	if err != nil {
		return err
	}

	d.Set("name", readApplication.Name)
	d.Set("label", readApplication.Label)
	d.Set("sign_on_mode", readApplication.SignOnMode)
	d.Set("aws_environment_type", readApplication.Settings.App.AwsEnvironmentType)
	d.Set("group_filter", readApplication.Settings.App.GroupFilter)
	d.Set("login_url", readApplication.Settings.App.LoginURL)
	d.Set("join_all_roles", readApplication.Settings.App.JoinAllRoles)
	d.Set("identity_provider_arn", readApplication.Settings.App.IdentityProviderArn)
	d.Set("session_duration", readApplication.Settings.App.SessionDuration)
	d.Set("role_value_pattern", readApplication.Settings.App.RoleValuePattern)
	d.Set("saml_metadata_document", samlMetadataDocument)

	fmt.Printf("%+v\n", readApplication)
	return nil
}

func applicationIsProvisioned(app *api.OktaApplication) error {
	for _, feat := range app.Features {
		if feat == "PUSH_NEW_USERS" {
			return nil
		}
	}
	return fmt.Errorf("PUSH_NEW_USERS is not configured")
}

func resourceAppAwsProvisionDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(Config)
	client := config.Okta
	web := config.Web
	appID := d.Id()

	err := try.Do(func(ampt int) (bool, error) {
		err := web.RevokeAWSProvisioning(appID)
		if err != nil {
			time.Sleep(RetryWaitTimeInSeconds)
			return ampt < client.RetryMaximum, err
		}

		app, err := client.GetApplication(appID)
		if err != nil {
			time.Sleep(RetryWaitTimeInSeconds)
			return ampt < client.RetryMaximum, err
		}

		err = applicationIsProvisioned(app)
		if err == nil {
			time.Sleep(RetryWaitTimeInSeconds)
			return ampt < client.RetryMaximum, err
		}

		return ampt < client.RetryMaximum, nil
	})

	return err
}

package okta

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

type Settings struct {
	App AppSettings `json:"app"`
}

type Application struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Label      string   `json:"label"`
	SignOnMode string   `json:"signOnMode"`
	Settings   Settings `json:"settings"`
}

type AppSettings struct {
	AwsEnvironmentType  string `json:"awsEnvironmentType"`
	GroupFilter         string `json:"groupFilter"`
	LoginURL            string `json:"loginUrl"`
	JoinAllRoles        bool   `json:"joinAllRoles"`
	SessionDuration     int    `json:"sessionDuration"`
	RoleValuePattern    string `json:"roleValuePattern"`
	IdentityProviderArn string `json:"identityProviderArn"`
}

type IdentifiedApplication struct {
	Application
	Credentials Credentials `json:"credentials"`
}

type Credentials struct {
	Signing Signing `json:"signing"`
}

type Signing struct {
	KeyID string `json:"kid"`
}

func resourceApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppCreate,
		Read:   resourceAppRead,
		Update: resourceAppUpdate,
		Delete: resourceAppDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"label": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"sign_on_mode": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "SAML_2_0",
			},
			"aws_environment_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"group_filter": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"login_url": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"join_all_roles": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"identity_provider_arn": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"session_duration": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3600,
			},
			"role_value_pattern": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"saml_metadata_document": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
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

func resourceAppCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	awsKey := d.Get("aws_okta_iam_user_id").(string)
	awsSecret := d.Get("aws_okta_iam_user_secret").(string)

	application := Application{
		Name:       d.Get("name").(string),
		Label:      d.Get("label").(string),
		SignOnMode: d.Get("sign_on_mode").(string),
		Settings: Settings{
			App: AppSettings{
				AwsEnvironmentType:  d.Get("aws_environment_type").(string),
				GroupFilter:         d.Get("group_filter").(string),
				LoginURL:            d.Get("login_url").(string),
				JoinAllRoles:        d.Get("join_all_roles").(bool),
				SessionDuration:     d.Get("session_duration").(int),
				RoleValuePattern:    d.Get("role_value_pattern").(string),
				IdentityProviderArn: d.Get("identity_provider_arn").(string),
			},
		},
	}

	createdApplication, err := client.CreateApplication(application)
	if err != nil {
		return err
	}

	samlMetadataDocument, err := client.GetSAMLMetaData(createdApplication.ID, createdApplication.Credentials.Signing.KeyID)
	if err != nil {
		return err
	}

	provisionErr := client.SetProvisioningSettings(createdApplication.ID, awsKey, awsSecret)
	if provisionErr != nil {
		return provisionErr
	}

	fmt.Printf("%+v\n", createdApplication)
	d.SetId(createdApplication.ID)
	d.Set("saml_metadata_document", samlMetadataDocument)

	return nil
}

func resourceAppRead(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	appID := d.Id()

	readApplication, applicationRemoved, err := client.ReadApplication(appID)
	if err != nil {
		return err
	}

	if applicationRemoved == true {
		log.Printf("[WARN] Okta Application %s (%q) not found, removing from state", d.Get("label").(string), d.Id())
		d.SetId("")
		return nil
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

	fmt.Printf("%+v\n", readApplication)
	return nil
}

func resourceAppUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	awsKey := d.Get("aws_okta_iam_user_id").(string)
	awsSecret := d.Get("aws_okta_iam_user_secret").(string)

	application := Application{
		ID:         d.Id(),
		Name:       d.Get("name").(string),
		Label:      d.Get("label").(string),
		SignOnMode: d.Get("sign_on_mode").(string),
		Settings: Settings{
			App: AppSettings{
				AwsEnvironmentType:  d.Get("aws_environment_type").(string),
				GroupFilter:         d.Get("group_filter").(string),
				LoginURL:            d.Get("login_url").(string),
				JoinAllRoles:        d.Get("join_all_roles").(bool),
				SessionDuration:     d.Get("session_duration").(int),
				RoleValuePattern:    d.Get("role_value_pattern").(string),
				IdentityProviderArn: d.Get("identity_provider_arn").(string),
			},
		},
	}

	updatedApplication, err := client.UpdateApplication(application)
	if err != nil {
		return err
	}

	provisionErr := client.SetProvisioningSettings(createdApplication.ID, awsKey, awsSecret)
	if provisionErr != nil {
		return provisionErr
	}

	fmt.Printf("%+v\n", updatedApplication)
	d.SetId(updatedApplication.ID)

	return nil
}

func resourceAppDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	appID := d.Id()

	err := client.DeleteApplication(appID)

	if err != nil {
		return err
	}

	return nil
}

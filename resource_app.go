package main

import (
	"fmt"

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
		},
	}
}

func resourceAppCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)

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

	fmt.Printf("%+v\n", createdApplication)
	d.SetId(createdApplication.ID)

	return nil
}

func resourceAppRead(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	appID := d.Id()

	readApplication, err := client.ReadApplication(appID)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", readApplication)
	return nil
}

func resourceAppUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)

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

	updatedApplication, err := client.UpdateApplication(application)
	if err != nil {
		return err
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

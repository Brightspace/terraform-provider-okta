package okta

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAppUsers() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppUsersCreate,
		Read:   resourceAppUsersRead,
		Update: resourceAppUsersUpdate,
		Delete: resourceAppUsersDelete,

		Schema: map[string]*schema.Schema{
			"app_id": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			//owners				
			"role_owner": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"user_owner": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			//users				
			"role_user": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"user_user": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			//readonly				
			"role_readonly": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"user_readonly": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			//finance				
			"role_finance": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"user_finance": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			//dashboard				
			"role_dashboard": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"user_dashboard": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			//dns				
			"role_dns": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"user_dns": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			//dns-admin				
			"role_dns_admin": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"user_dns_admin": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			//athena				
			"role_athena": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"user_athena": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			//athena-admin				
			"role_athena_admin": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"user_athena_admin": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
		},
	}
}

func resourceAppUsersCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	app_id := d.Get("app_id").(string)
	role := d.Get("role").(string)
	saml_roles := d.Get("saml_roles").([]interface{})
	roles := make([]string, len(saml_roles))
	for i, value := range saml_roles {
		roles[i] = value.(string)
	}

	user_id, err := client.GetUserIDByEmail(d.Get("user").(string))
	if err != nil {
		return err
	}

	_, err = client.AddMemberToApp(app_id, user_id, role, roles)
	if err != nil {
		return err
	}

	d.SetId(user_id)

	return resourceAppUsersRead(d, m)
}

func resourceAppUsersUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	app_id := d.Get("app_id").(string)
	role := d.Get("role").(string)
	saml_roles := d.Get("saml_roles").([]interface{})
	roles := make([]string, len(saml_roles))
	for i, value := range saml_roles {
		roles[i] = value.(string)
	}

	_, err := client.AddMemberToApp(app_id, d.Id(), role, roles)
	if err != nil {
		return err
	}

	return resourceAppUsersRead(d, m)
}

func resourceAppUsersRead(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)

	member, err := client.GetAppMember(d.Get("app_id").(string), d.Id())
	if err != nil {
		log.Printf("[WARN] User (%s) in app (%s) not found, removing from state", d.Id(), d.Get("app_id").(string))
		d.SetId("")
		return nil
	}

	log.Printf("[INFO] App %s user (%s) discovered", d.Get("app_id").(string), d.Id())

	d.Set("status", member.Status)
	d.Set("email", member.Profile.Email)
	d.Set("display_name", member.Profile.DisplayName)
	d.Set("role", member.Profile.Role)
	d.Set("saml_roles", member.Profile.SamlRoles)

	return nil
}

func resourceAppUsersDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)

	err := client.RemoveMemberFromApp(d.Get("app_id").(string), d.Id())
	if err != nil {
		return err
	}

	return nil
}

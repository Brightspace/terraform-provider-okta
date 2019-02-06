package okta

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAppUserAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppUserAttachmentCreate,
		Read:   resourceAppUserAttachmentRead,
		Update: resourceAppUserAttachmentUpdate,
		Delete: resourceAppUserAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"role": &schema.Schema{
				Type: schema.TypeString,
			},
			"user": &schema.Schema{
				Type: schema.TypeString,
			},
			"app_id": &schema.Schema{
				Type: schema.TypeString,
			},
			"saml_roles": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceAppUserAttachmentCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	user_id := d.Get("user").(string)
	app_id := d.Get("app_id").(string)
	role := d.Get("role").(string)
	saml_roles := d.Get("saml_roles").([]string)

	_, err := client.AddMemberToApp(app_id, user_id, role, saml_roles)
	if err != nil {
		return err
	}

	d.SetId(user_id)

	return resourceAppUserAttachmentRead(d, m)
}

func resourceAppUserAttachmentUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	app_id := d.Get("app_id").(string)
	role := d.Get("role").(string)
	saml_roles := d.Get("saml_roles").([]string)

	_, err := client.AddMemberToApp(app_id, d.Id(), role, saml_roles)
	if err != nil {
		return err
	}

	return resourceAppUserAttachmentRead(d, m)
}

func resourceAppUserAttachmentRead(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)

	member, err := client.GetAppMember(d.Get("app_id").(string), d.Id())
	if err != nil {
		return err
	}

	log.Printf("[INFO] App %s user (%s) discovered", d.Get("app_id").(string), d.Id())

	d.Set("status", member.Status)
	d.Set("email", member.Profile.Email)
	d.Set("display_name", member.Profile.DisplayName)
	d.Set("role", member.Profile.Role)
	d.Set("saml_roles", member.Profile.SamlRoles)

	return nil
}

func resourceAppUserAttachmentDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)

	err := client.RemoveMemberFromApp(d.Get("app_id").(string), d.Id())
	if err != nil {
		return err
	}

	return nil
}

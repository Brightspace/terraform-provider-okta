package okta

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAppUserAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppUserAttachmentCreate,
		Read:   resourceAppUserAttachmentRead,
		Delete: resourceAppUserAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"role": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"user": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"app_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"saml_roles": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
		},
	}
}

func resourceAppUserAttachmentCreate(d *schema.ResourceData, m interface{}) error {
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

	return resourceAppUserAttachmentRead(d, m)
}

func resourceAppUserAttachmentRead(d *schema.ResourceData, m interface{}) error {
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

func resourceAppUserAttachmentDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)

	err := client.RemoveMemberFromApp(d.Get("app_id").(string), d.Id())
	if err != nil {
		return err
	}

	return nil
}

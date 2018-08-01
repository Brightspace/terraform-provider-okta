package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAppGroupMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppGroupMembershipCreate,
		Read:   resourceAppGroupMembershipRead,
		Update: resourceAppGroupMembershipUpdate,
		Delete: resourceAppGroupMembershipDelete,

		Schema: map[string]*schema.Schema{
			"group_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"user": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "User email address e.g. dr.dre@example.com",
			},
		},
	}
}

func resourceAppGroupMembershipCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	user := d.Get("user").(string)
	groupID := d.Get("group_id").(string)

	err := client.AddMemeberToGroup(groupID, user)
	if err != nil {
		return err
	}

	return nil
}

func resourceAppGroupMembershipUpdate(d *schema.ResourceData, m interface{}) error {
	//client := m.(OktaClient)

	return nil
}

func resourceAppGroupMembershipRead(d *schema.ResourceData, m interface{}) error {
	//client := m.(OktaClient)

	return nil
}

func resourceAppGroupMembershipDelete(d *schema.ResourceData, m interface{}) error {
	//client := m.(OktaClient)

	return nil
}

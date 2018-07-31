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
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
		},
	}
}

func resourceAppGroupMembershipCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)

	return nil
}

func resourceAppGroupMembershipUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)

	return nil
}

func resourceAppGroupMembershipRead(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)

	return nil
}

func resourceAppGroupMembershipDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)

	return nil
}

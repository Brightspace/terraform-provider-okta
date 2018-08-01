package main

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAppGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppGroupCreate,
		Read:   resourceAppGroupRead,
		Update: resourceAppGroupUpdate,
		Delete: resourceAppGroupDelete,

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
			"members": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	}
}

func resourceAppGroupCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	groupMembers := d.Get("members").([]string)

	groupID, err := client.CreateGroup(d.Get("name").(string), d.Get("description").(string))
	if err != nil {
		return err
	}

	d.SetId(groupID)

	for _, member := range groupMembers {
		err := client.AddMemeberToGroup(groupID, member)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceAppGroupUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)

	err := client.UpdateGroup(d.Get("id").(string), d.Get("name").(string), d.Get("description").(string))
	if err != nil {
		return err
	}

	return nil
}

func resourceAppGroupRead(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)

	group, err := client.GetGroup(d.Get("id").(string))
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", group)
	return nil
}

func resourceAppGroupDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)

	err := client.DeleteGroup(d.Get("id").(string))
	if err != nil {
		return err
	}

	return nil
}

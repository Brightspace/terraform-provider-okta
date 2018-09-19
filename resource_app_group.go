package main

import (
	"fmt"
	"time"

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
			"app_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"saml_role": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
		},
	}
}

func resourceAppGroupCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	groupMembers := d.Get("members").([]interface{})

	groupID, err := client.CreateGroup(d.Get("name").(string), d.Get("description").(string))
	if err != nil {
		return err
	}

	members := make([]string, len(groupMembers))
	for i, member := range groupMembers {
		members[i] = member.(string)
	}

	err2 := client.SyncUsersToGroup(groupID, members)
	if err2 != nil {
		return err2
	}

	time.Sleep(10 * time.Second)
	err3 := client.AssignGroupToApp(d.Get("app_id").(string), groupID, d.Get("saml_role").(string))
	if err3 != nil {
		return err3
	}

	d.SetId(groupID)

	return nil
}

func resourceAppGroupUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	groupMembers := d.Get("members").([]interface{})

	err := client.UpdateGroup(d.Id(), d.Get("name").(string), d.Get("description").(string))
	if err != nil {
		return err
	}

	members := make([]string, len(groupMembers))
	for i, member := range groupMembers {
		members[i] = member.(string)
	}

	err2 := client.SyncUsersToGroup(d.Id(), members)
	if err2 != nil {
		return err2
	}

	return nil
}

func resourceAppGroupRead(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)

	group, groupMembers, groupRemoved, err := client.GetGroup(d.Id())
	if err != nil {
		return err
	}

	if groupRemoved == true {
		d.SetId("")
		return nil
	}

	members := make([]interface{}, len(groupMembers))
	for i, groupMember := range groupMembers {
		members[i] = groupMember.Profile.Login
	}

	d.Set("name", group.Profile.Name)
	d.Set("description", group.Profile.Description)
	d.Set("members", members)

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

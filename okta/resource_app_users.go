package okta

import (
	"log"
	"sort"
	"strings"

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

func composeSamlMapping(mappings map[string][]string) map[string][]string {
	result := make(map[string][]string)
	for saml, users := range mappings {
		for _, user := range users {
			if _, exists := result[user]; !exists {
				result[user] = []string{saml}
			} else {
				result[user] = append(result[user], saml)
			}
		}
	}

	return result
}

func convertToStrings(d *schema.ResourceData, property string) []string {
	arrays := d.Get(property).([]interface{})
	result := make([]string, len(arrays))
	for i, value := range arrays {
		result[i] = value.(string)
	}
	return result
}

func composeRoleMappings(d *schema.ResourceData) map[string][]string {
	mapping := map[string][]string{
		d.Get("role_owner").(string):        convertToStrings(d, "user_owner"),
		d.Get("role_user").(string):         convertToStrings(d, "user_user"),
		d.Get("role_readonly").(string):     convertToStrings(d, "user_readonly"),
		d.Get("role_finance").(string):      convertToStrings(d, "user_finance"),
		d.Get("role_dashboard").(string):    convertToStrings(d, "user_dashboard"),
		d.Get("role_dns").(string):          convertToStrings(d, "user_dns"),
		d.Get("role_dns_admin").(string):    convertToStrings(d, "user_dns_admin"),
		d.Get("role_athena").(string):       convertToStrings(d, "user_athena"),
		d.Get("role_athena_admin").(string): convertToStrings(d, "user_athena_admin"),
	}
	return mapping
}

func arraysEqual(x []string, y []string) bool {
	xmap := make(map[string]int)
	ymap := make(map[string]int)

	for _, item := range x {
		xmap[item]++
	}
	for _, item := range y {
		ymap[item]++
	}

	for k, v := range xmap {
		if ymap[k] != v {
			return false
		}
	}
	return true
}

func composeResourceId(users []string) string {
	sort.Strings(users)
	return strings.Join(users, "+")
}

func resourceAppUsersCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	app_id := d.Get("app_id").(string)
	role := d.Get("role_readonly").(string)

	role_mapping := composeRoleMappings(d)
	saml_mapping := composeSamlMapping(role_mapping)
	var ids []string

	for user, roles := range saml_mapping {
		user_id, err := client.GetUserIDByEmail(user)
		if err != nil {
			return err
		}

		_, err = client.AddMemberToApp(app_id, user_id, role, roles)
		if err != nil {
			return err
		}

		ids = append(ids, user_id)
	}

	d.SetId(composeResourceId(ids))

	return resourceAppUsersRead(d, m)
}

func resourceAppUsersUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	app_id := d.Get("app_id").(string)
	role := d.Get("role_readonly").(string)

	role_mapping := composeRoleMappings(d)
	saml_mapping := composeSamlMapping(role_mapping)
	var ids []string

	for user, roles := range saml_mapping {
		user_id, err := client.GetUserIDByEmail(user)
		if err != nil {
			return err
		}

		ids = append(ids, user_id)

		update := false

		member, err := client.GetAppMember(app_id, user_id)
		if err != nil {
			log.Printf("[WARN] User (%s) in app (%s) not found, removing from state", d.Id(), app_id)
			update = true
		}

		if arraysEqual(member.Profile.SamlRoles, roles) {
			update = false
		}

		if update {
			_, err := client.AddMemberToApp(app_id, user_id, role, roles)
			if err != nil {
				return err
			}
		}
	}

	members, err := client.ListAppMembers(app_id)
	if err != nil {
		return err
	}

	for _, member := range members {
		if _, exists := saml_mapping[member.Profile.Email]; !exists {
			continue
		}

		err := client.RemoveMemberFromApp(d.Get("app_id").(string), member.ID)
		if err != nil {
			return err
		}
	}

	return resourceAppUsersRead(d, m)
}

func resourceAppUsersRead(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	app_id := d.Get("app_id").(string)
	role_mapping := composeRoleMappings(d)
	saml_mapping := composeSamlMapping(role_mapping)

	members, err := client.ListAppMembers(app_id)
	if err != nil {
		return err
	}

	var ids []string
	for user, roles := range saml_mapping {
		user_id, err := client.GetUserIDByEmail(user)
		if err != nil {
			return err
		}

		for _, member := range members {
			if !strings.EqualFold(user, member.Profile.Email) {
				continue
			}

			if !arraysEqual(roles, member.Profile.SamlRoles) {
				log.Printf("[WARN] User (%s) in app (%s) does not have all SAML roles", user, app_id)
				continue
			}

			ids = append(ids, member.ID)
		}

		if !strings.Contains(d.Id(), user_id) {
			log.Printf("[WARN] User (%s) in app (%s) not found, removing from state", user, app_id)
		}
	}

	d.SetId(composeResourceId(ids))

	return nil
}

func resourceAppUsersDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	app_id := d.Get("app_id").(string)

	members, err := client.ListAppMembers(app_id)
	if err != nil {
		return err
	}

	for _, member := range members {
		err := client.RemoveMemberFromApp(d.Get("app_id").(string), member.ID)
		if err != nil {
			return err
		}
	}

	d.SetId("")

	return nil
}

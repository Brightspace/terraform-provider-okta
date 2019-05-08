package okta

import (
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"sort"
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
			//storage
			"role_storage": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"user_storage": &schema.Schema{
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

func getRoleMappings(d *schema.ResourceData) map[string]string {
	mapping := map[string]string{
		d.Get("role_owner").(string):        "user_owner",
		d.Get("role_user").(string):         "user_user",
		d.Get("role_readonly").(string):     "user_readonly",
		d.Get("role_finance").(string):      "user_finance",
		d.Get("role_dashboard").(string):    "user_dashboard",
		d.Get("role_storage").(string):      "user_storage",
		d.Get("role_dns").(string):          "user_dns",
		d.Get("role_dns_admin").(string):    "user_dns_admin",
		d.Get("role_athena").(string):       "user_athena",
		d.Get("role_athena_admin").(string): "user_athena_admin",
	}
	return mapping
}

func composeRoleMappings(d *schema.ResourceData) map[string][]string {
	mapping := map[string][]string{
		d.Get("role_owner").(string):        convertToStrings(d, "user_owner"),
		d.Get("role_user").(string):         convertToStrings(d, "user_user"),
		d.Get("role_readonly").(string):     convertToStrings(d, "user_readonly"),
		d.Get("role_finance").(string):      convertToStrings(d, "user_finance"),
		d.Get("role_dashboard").(string):    convertToStrings(d, "user_dashboard"),
		d.Get("role_storage").(string):      convertToStrings(d, "user_storage"),
		d.Get("role_dns").(string):          convertToStrings(d, "user_dns"),
		d.Get("role_dns_admin").(string):    convertToStrings(d, "user_dns_admin"),
		d.Get("role_athena").(string):       convertToStrings(d, "user_athena"),
		d.Get("role_athena_admin").(string): convertToStrings(d, "user_athena_admin"),
	}
	return mapping
}

func arraysEqual(x []string, y []string) bool {
	if len(x) != len(y) {
		return false
	}

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

	d.SetId(app_id)

	return resourceAppUsersRead(d, m)
}

func resourceAppUsersUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	app_id := d.Get("app_id").(string)
	role := d.Get("role_readonly").(string)

	role_mapping := composeRoleMappings(d)
	saml_mapping := composeSamlMapping(role_mapping)
	ids := make(map[string]bool)

	for user, roles := range saml_mapping {
		user_id, err := client.GetUserIDByEmail(user)
		if err != nil {
			return err
		}
		log.Printf("[INFO] User (%s) found as %s to roles %s", user, user_id, roles)
		ids[user_id] = true

		update := true
		member, err := client.GetAppMember(app_id, user_id)
		if err != nil || member.ID == "" {
			log.Printf("[WARN] User (%s) in app (%s) not found, re-adding", user, app_id)
			update = true
		} else if arraysEqual(roles, member.Profile.SamlRoles) {
			log.Printf("[INFO] User (%s) in app (%s) matches roles of %s", user, app_id, member.Profile.SamlRoles)
			update = false
		}

		if update {
			log.Printf("[WARN] Updating user (%s) in app (%s) to fit roles", user, app_id)
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
		if _, exists := ids[member.ID]; exists {
			continue
		}

		log.Printf("[WARN] User (%s) in app (%s) not found in list, removing", member.Profile.Email, app_id)
		err := client.RemoveMemberFromApp(app_id, member.ID)
		if err != nil {
			return err
		}
	}

	return resourceAppUsersRead(d, m)
}

func resourceAppUsersRead(d *schema.ResourceData, m interface{}) error {
	client := m.(OktaClient)
	app_id := d.Get("app_id").(string)
	samlToUser := getRoleMappings(d)

	members, err := client.ListAppMembers(app_id)
	if err != nil {
		return err
	}

	existing := make(map[string][]string)
	for saml, _ := range samlToUser {
		existing[saml] = []string{}
	}

	for _, member := range members {
		log.Printf("[INFO] Reading user (%s) on app %s", member.ID, app_id)
		user, err := client.GetAppMember(app_id, member.ID)
		if err != nil {
			log.Printf("[ERR] Failed to read user (%s) from app %s", member.ID, app_id)
			return err
		}

		log.Printf("[INFO] Reading user (%s) with roles: %s", member.ID, user.Profile.SamlRoles)
		for _, saml := range user.Profile.SamlRoles {
			user := user.Profile.Email
			log.Printf("[INFO] Identifying user (%s) on for role %s", user, saml)

			existing[saml] = append(existing[saml], user)
		}
	}

	for saml, users := range existing {
		out_users := samlToUser[saml]
		sort.Strings(users)
		log.Printf("[INFO] Setting (%s) role %s to: %s", out_users, saml, users)
		d.Set(out_users, users)
	}

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
		err := client.RemoveMemberFromApp(app_id, member.ID)
		if err != nil {
			return err
		}
	}

	d.SetId("")

	return nil
}

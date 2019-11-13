package main

import (
	"fmt"
	"github.com/Brightspace/terraform-provider-okta/okta/api"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Missing ID for okta resource")
	}
	appId := os.Args[1]
	email := os.Args[2]

	client := api.Okta{
		APIKey:       os.Getenv("OKTA_API_KEY"),
		HostURL:      os.Getenv("OKTA_URL"),
		OrgID:        os.Getenv("OKTA_ORG_ID"),
		RetryMaximum: 5,
	}

	userId, err := client.GetUserIDByEmail(email, "")
	if err != nil {
		fmt.Println("err:\n", err)
		return
	}

	result, err := client.GetAppMember(appId, userId)
	if result == nil {
		fmt.Println("id could not be found:\n", appId)
		return
	}

	if err != nil {
		fmt.Println("err:\n", err)
		return
	}

	fmt.Println("ID:\n", result.ID)
	fmt.Println("Username:\n", result.Profile.Email)
}

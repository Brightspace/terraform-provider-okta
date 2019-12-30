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
	userId := os.Args[2]

	client := api.Okta{
		APIKey:       os.Getenv("OKTA_API_KEY"),
		HostURL:      os.Getenv("OKTA_URL"),
		OrgID:        os.Getenv("OKTA_ORG_ID"),
		RetryMaximum: 5,
	}

	err := client.RemoveAppMember(appId, userId)
	if err != nil {
		fmt.Println("err:\n", err)
		return
	}

	result, _ := client.ListAppMembers(appId)
	for _, element := range result {
		fmt.Println("ID:\n", element.ID)
		fmt.Println("Username:\n", element.Profile.Email)
	}
}

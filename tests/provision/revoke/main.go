package main

import (
	"fmt"
	"github.com/Brightspace/terraform-provider-okta/okta/api"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing required inputs")
	}

	appId := os.Args[1]
	client := api.OktaWebClient{
		UserName: os.Getenv("OKTA_USERNAME"),
		Password: os.Getenv("OKTA_PASSWORD"),
		AdminURL: os.Getenv("OKTA_ADMIN_URL"),
		HostURL:  os.Getenv("OKTA_URL"),
		OrgID:    os.Getenv("OKTA_ORG_ID"),
	}
	err := client.RevokeAWSProvisioning(appId)
	if err != nil {
		fmt.Println("err:\n", err)
		return
	}
}

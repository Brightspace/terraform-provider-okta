package main

import (
	"fmt"
	"github.com/Brightspace/terraform-provider-okta/okta/api"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing ID for evident resource")
	}
	appId := os.Args[1]

	client := api.Okta{
		APIKey:       os.Getenv("OKTA_API_KEY"),
		HostURL:      os.Getenv("OKTA_HOST_URL"),
		OrgID:        os.Getenv("OKTA_ORG_ID"),
		RetryMaximum: 5,
	}

	result, err := client.GetApp(appId)
	if result == nil {
		fmt.Println("id could not be found:\n", appId)
		return
	}

	if err != nil {
		fmt.Println("err:\n", err)
		return
	}
	fmt.Println("id:\n", result.ID)
	fmt.Println("name:\n", result.Attributes.Name)
	fmt.Println("arn:\n", result.Attributes.Arn)
	fmt.Println("external_id:\n", result.Attributes.ExternalID)
}

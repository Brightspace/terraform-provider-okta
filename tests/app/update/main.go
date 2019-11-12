package main

import (
	"fmt"
	"github.com/Brightspace/terraform-provider-okta/okta/api"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Missing required inputs")
	}

	appId := os.Args[1]
	name := os.Args[2]

	client := api.Okta{
		APIKey:       os.Getenv("OKTA_API_KEY"),
		HostURL:      os.Getenv("OKTA_HOST_URL"),
		OrgID:        os.Getenv("OKTA_ORG_ID"),
		RetryMaximum: 5,
	}

	app, err := client.GetApplication(appId)
	if app == nil {
		fmt.Println("id could not be found:\n", appId)
		return
	}
	
	if err != nil {
		fmt.Println("err:\n", err)
		return
	}

	result, err := client.UpdateAwsApplication(app.ID, name, app.Settings.App.IdentityProviderArn)
	if err != nil {
		fmt.Println("err:\n", err)
		return
	}

	fmt.Println("ID:\n", result.ID)
	fmt.Println("Name:\n", result.Name)
	fmt.Println("Label:\n", result.Label)
	fmt.Println("SignOnMode:\n", result.SignOnMode)
	fmt.Println("Signing.KeyID:\n", result.Credentials.Signing.KeyID)
}

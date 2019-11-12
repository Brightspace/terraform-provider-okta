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
	name := os.Args[1]
	arn := os.Args[2]

	client := api.Okta{
		APIKey:       os.Getenv("OKTA_API_KEY"),
		HostURL:      os.Getenv("OKTA_HOST_URL"),
		OrgID:        os.Getenv("OKTA_ORG_ID"),
		RetryMaximum: 5,
	}

	result, err := client.CreateAwsApplication(name, arn)
	if err != nil {
		fmt.Println("Error:\n", err)
		return
	}

	if result == nil {
		fmt.Println("Nothing returned")
		return
	}

	fmt.Println("ID:\n", result.ID)
	fmt.Println("Name:\n", result.Name)
	fmt.Println("Label:\n", result.Label)
	fmt.Println("SignOnMode:\n", result.SignOnMode)
}

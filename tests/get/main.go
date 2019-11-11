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
	arg := os.Args[1]

	client := api.Okta{
		Credentials: api.Credentials{
			AccessKey: []byte(os.Getenv("EVIDENT_ACCESS_KEY")),
			SecretKey: []byte(os.Getenv("EVIDENT_SECRET_KEY")),
		},
		RetryMaximum: 5,
	}

	result, err := client.Get(arg)
	if result == nil {
		fmt.Println("id could not be found:\n", arg)
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

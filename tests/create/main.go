package main

import (
	"fmt"
	"github.com/Brightspace/terraform-provider-evident/evident/api"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 4 {
		log.Fatal("Missing required inputs")
	}
	name := "EvidentProviderTest"
	arn := os.Args[1]
	team_id := os.Args[2]
	external_id := os.Args[3]

	client := api.Evident{
		Credentials: api.Credentials{
			AccessKey: []byte(os.Getenv("EVIDENT_ACCESS_KEY")),
			SecretKey: []byte(os.Getenv("EVIDENT_SECRET_KEY")),
		},
		RetryMaximum: 5,
	}

	result, _ := client.Add(name, arn, external_id, team_id)
	fmt.Println("id:\n", result.ID)
	fmt.Println("name:\n", result.Attributes.Name)
	fmt.Println("arn:\n", result.Attributes.Arn)
	fmt.Println("external_id:\n", result.Attributes.ExternalID)
}

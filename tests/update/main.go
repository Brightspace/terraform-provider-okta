package main

import (
	"fmt"
	"github.com/Brightspace/terraform-provider-evident/evident/api"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 5 {
		log.Fatal("Missing required inputs")
	}
	name := "EvidentProviderTestUpdate"
	id := os.Args[1]
	arn := os.Args[2]
	team_id := os.Args[3]
	external_id := os.Args[4]

	client := api.Evident{
		Credentials: api.Credentials{
			AccessKey: []byte(os.Getenv("EVIDENT_ACCESS_KEY")),
			SecretKey: []byte(os.Getenv("EVIDENT_SECRET_KEY")),
		},
		RetryMaximum: 5,
	}

	result, err := client.Update(id, name, arn, external_id, team_id)
	fmt.Println("err:\n", err)
	fmt.Println("id:\n", result.ID)
	fmt.Println("name:\n", result.Attributes.Name)
	fmt.Println("arn:\n", result.Attributes.Arn)
	fmt.Println("external_id:\n", result.Attributes.ExternalID)
}

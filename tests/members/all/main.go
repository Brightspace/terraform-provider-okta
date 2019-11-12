package main

import (
	"fmt"
	"github.com/Brightspace/terraform-provider-evident/evident/api"
	"os"
)

func main() {
	client := api.Evident{
		Credentials: api.Credentials{
			AccessKey: []byte(os.Getenv("EVIDENT_ACCESS_KEY")),
			SecretKey: []byte(os.Getenv("EVIDENT_SECRET_KEY")),
		},
		RetryMaximum: 5,
	}

	result, _ := client.All()
	for _, element := range result {
		fmt.Println("name:\n", element.Attributes.Name)
	}
}

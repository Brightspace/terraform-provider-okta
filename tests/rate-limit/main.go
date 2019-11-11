package main

import (
	"fmt"
	"github.com/Brightspace/terraform-provider-evident/evident/api"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing ID for evident resource")
	}
	id := os.Args[1]
	counter := 10

	client := api.Evident{
		Credentials: api.Credentials{
			AccessKey: []byte(os.Getenv("EVIDENT_ACCESS_KEY")),
			SecretKey: []byte(os.Getenv("EVIDENT_SECRET_KEY")),
		},
		RetryMaximum: 5,
	}

	for i := 0; i < counter; i++ {
		result, err := client.Get(id)
		if err != nil {
			fmt.Println("Encountered an error:\n", err)
		} else if result == nil {
			fmt.Println("404 not found")
		} else if result.ID != id {
			fmt.Println("ID mismatch:\n", result.ID)
		} else {
			// Ignore sometimes
			fmt.Println("ID:\n", result.ID)
		}
	}
}

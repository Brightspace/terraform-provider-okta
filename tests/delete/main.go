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
	arg := os.Args[1]

	client := api.Evident{
		Credentials: api.Credentials{
			AccessKey: []byte(os.Getenv("EVIDENT_ACCESS_KEY")),
			SecretKey: []byte(os.Getenv("EVIDENT_SECRET_KEY")),
		},
		RetryMaximum: 5,
	}

	resp, _ := client.Delete(arg)
	fmt.Println(resp)

	result, _ := client.Get(arg)
	if result != nil {
		fmt.Println("success:\n", arg)
	} else {
		fmt.Println("fail:\n", arg)
	}
}

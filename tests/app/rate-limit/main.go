package main

import (
	"fmt"
	"github.com/Brightspace/terraform-provider-okta/okta/api"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing ID for evident resource")
	}
	appId := os.Args[1]
	counter := 100

	client := api.Okta{
		APIKey:       os.Getenv("OKTA_API_KEY"),
		HostURL:      os.Getenv("OKTA_URL"),
		OrgID:        os.Getenv("OKTA_ORG_ID"),
		RetryMaximum: 5,
	}

	result, err := client.GetApplication(appId)
	if result == nil {
		fmt.Println("id could not be found:\n", appId)
		return
	}

	if err != nil {
		fmt.Println("err:\n", err)
		return
	}

	fmt.Println("ID:\n", result.ID)
	fmt.Println("KeyID:\n", result.Credentials.Signing.KeyID)

	for i := 0; i < counter; i++ {
		samlMetaData, err := client.GetSAMLMetadata(result.ID, result.Credentials.Signing.KeyID)
		if err != nil {
			fmt.Println("Encountered an error:\n", err)
		} else if samlMetaData == "" {
			fmt.Println("nothing returned")
		} else if strings.Contains(samlMetaData, "E0000047") {
			fmt.Println("Rate limit hit:\n", i)
		}
	}
}

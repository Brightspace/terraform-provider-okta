package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Config struct {
	OktaURL string
	APIKey  string
}

type OktaClient struct {
	OktaURL string
	APIKey  string
}

func NewClient(c *Config) OktaClient {
	client = initHTTPClient()

	return OktaClient{
		OktaURL: c.OktaURL,
		APIKey:  c.APIKey,
	}
}

var client http.Client
var config *Config

func initHTTPClient() http.Client {
	timeout := time.Duration(time.Second * 30)
	return http.Client{
		Timeout: timeout,
	}
}

func (o *OktaClient) CreateApplication(application Application) (Application, error) {
	var app Application
	url := fmt.Sprintf("%s/api/v1/apps", o.OktaURL)

	body, err := json.Marshal(application)
	if err != nil {
		return app, err
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	res, err := client.Do(req)

	if err != nil {
		return app, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		msg := buf.String()

		return app, fmt.Errorf("Error creating application in Okta: %s", msg)
	}

	err = json.NewDecoder(res.Body).Decode(&app)
	if err != nil {
		return app, err
	}

	return app, nil
}

func (o *OktaClient) DeleteApplication(appID string) error {
	// Deactivate app first
	url := fmt.Sprintf("%s/api/v1/apps/%s/lifecycle/deactivate", o.OktaURL, appID)

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	res, err := client.Do(req)

	if err != nil {
		return err
	}

	// Delete app
	url = fmt.Sprintf("%s/api/v1/apps/%s", o.OktaURL, appID)

	req, _ = http.NewRequest("DELETE", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	res, err = client.Do(req)

	if err != nil {
		return err
	}

	if res.StatusCode != 204 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		msg := buf.String()

		return fmt.Errorf("Error deleting application in Okta: %s", msg)
	}

	return nil
}

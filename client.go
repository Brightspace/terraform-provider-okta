package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Config struct {
	OktaURL      string
	OktaAdminUrl string
	APIKey       string
	UserName     string
	Password     string
	OrgID        string
}

type OktaClient struct {
	OktaURL      string
	OktaAdminUrl string
	APIKey       string
	UserName     string
	Password     string
	OrgID        string
}

type OktaAuthResponse struct {
	ExpiresAt    time.Time `json:"expiresAt"`
	SessionToken string    `json:"sessionToken"`
	StateToken   string    `json:"stateToken"`
	Status       string    `json:"status"`
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

func (o *OktaClient) CreateApplication(application Application) (IdentifiedApplication, error) {
	var idApp IdentifiedApplication

	url := fmt.Sprintf("%s/api/v1/apps", o.OktaURL)

	body, err := json.Marshal(application)
	if err != nil {
		return idApp, err
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	res, err := client.Do(req)

	if err != nil {
		return idApp, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		msg := buf.String()

		return idApp, fmt.Errorf("Error creating application in Okta: %s", msg)
	}

	err = json.NewDecoder(res.Body).Decode(&idApp)
	if err != nil {
		return idApp, err
	}

	return idApp, nil
}

func (o *OktaClient) UpdateApplication(application Application) (Application, error) {
	var app Application
	url := fmt.Sprintf("%s/api/v1/apps/%s", o.OktaURL, app.ID)

	body, err := json.Marshal(application)
	if err != nil {
		return app, err
	}

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
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

		return app, fmt.Errorf("Error updating application in Okta: %s", msg)
	}

	err = json.NewDecoder(res.Body).Decode(&app)
	if err != nil {
		return app, err
	}

	return app, nil
}

func (o *OktaClient) ReadApplication(appID string) (Application, error) {
	var app Application
	url := fmt.Sprintf("%s/api/v1/apps/%s", o.OktaURL, appID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	res, err := client.Do(req)
	if err != nil {
		return app, err
	}

	err = json.NewDecoder(res.Body).Decode(&app)
	if err != nil {
		return app, err
	}

	return app, err
}

func (o *OktaClient) GetSAMLMetaData(appID string, keyID string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/apps/%s/sso/saml/metadata?kid=%s", o.OktaURL, appID, keyID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	samlMetaData := buf.String()

	return samlMetaData, nil
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

func (o *OktaClient) SetProvisioningSettings(appID string, oktaAWSKey string, oktaAWSSecretKey string) error {
	authBody := fmt.Sprintf(`{"username":"%s", "password":"%s"}`, o.UserName, o.Password)

	cookieJar, _ := cookiejar.New(nil)
	client.Jar = cookieJar

	authUrl := fmt.Sprintf(`%s/api/v1/authn`, o.OktaURL)
	req, _ := http.NewRequest("POST", authUrl, strings.NewReader(authBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	oktaAuthResponse := &OktaAuthResponse{}
	body, err := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &oktaAuthResponse)

	cookieUrl := fmt.Sprintf("%s/login/sessionCookieRedirect?checkAccountSetupComplete=true&token=%s&redirectUrl=%s/user/notifications", o.OktaURL, oktaAuthResponse.SessionToken, o.OktaURL)

	req2, _ := http.NewRequest("GET", cookieUrl, nil)
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Accept", "application/json")

	_, err2 := client.Do(req2)
	if err2 != nil {
		return err2
	}

	// ---------------
	userHomeUrl := fmt.Sprintf("%s/app/UserHome", o.OktaURL)
	req3, _ := http.NewRequest("GET", userHomeUrl, nil)
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Accept", "application/json")

	_, err3 := client.Do(req3)
	if err3 != nil {
		return err3
	}

	// ---------------
	oneUrl := fmt.Sprintf("%s/home/saasure/%s/1", o.OktaURL, o.OrgID)
	req5, _ := http.NewRequest("GET", oneUrl, nil)
	req5.Header.Set("Content-Type", "application/json")
	req5.Header.Set("Accept", "application/json")

	oneResp, err5 := client.Do(req5)
	if err5 != nil {
		return err5
	}

	ssoToken := getSsoToken(*oneResp)
	defer oneResp.Body.Close()

	// ---------------
	adminSsoUrl := fmt.Sprintf("%s/admin/sso/request", o.OktaAdminUrl)
	postData := url.Values{}
	postData.Add("token", ssoToken)
	req6, _ := http.NewRequest("POST", adminSsoUrl, strings.NewReader(postData.Encode()))
	req6.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req6.Header.Set("Accept", "application/json")

	ssoResp, err6 := client.Do(req6)
	if err6 != nil {
		return err6
	}

	xsrfToken := getXsrfToken(*ssoResp)
	defer ssoResp.Body.Close()

	// ---------------
	appUpdateUrl := fmt.Sprintf("%s/admin/app/amazon_aws/instance/%s/settings/user-mgmt", o.OktaAdminUrl, appID)
	updateAppData := url.Values{}
	updateAppData.Add("_xsrfToken", xsrfToken)
	updateAppData.Add("enabled", "true")
	updateAppData.Add("_enabled", "on")
	updateAppData.Add("accessKeyUM", oktaAWSKey)
	updateAppData.Add("secretKeyUM", oktaAWSSecretKey)
	updateAppData.Add("accountIds", "")
	updateAppData.Add("_pushNewAccount", "on")
	updateAppData.Add("_pushProfile", "on")
	updateAppData.Add("overrideApiURL", "")
	updateAppData.Add("pushNewAccount", "true")

	req7, _ := http.NewRequest("POST", appUpdateUrl, strings.NewReader(updateAppData.Encode()))
	req7.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req7.Header.Set("Accept", "application/json")

	_, err7 := client.Do(req7)
	if err7 != nil {
		return err7
	}

	return nil
}

func getXsrfToken(res http.Response) string {
	z := html.NewTokenizer(res.Body)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			return ""
		case tt == html.StartTagToken:
			t := z.Token()

			if t.Data == "span" {
				for _, a := range t.Attr {
					if a.Key == "id" && a.Val == "_xsrfToken" {
						z.Next()
						return z.Token().Data
					}
				}
			}
		}
	}
}

func getSsoToken(res http.Response) string {
	z := html.NewTokenizer(res.Body)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			return ""
		case tt == html.StartTagToken:
			t := z.Token()

			if t.Data == "script" {
				for _, a := range t.Attr {
					if a.Key == "type" && a.Val == "text/javascript" {
						z.Next()
						javascriptText := z.Token().Data
						r := regexp.MustCompile(`(?m)var\s*repostParams\s*=\s*{\s*"token"\s*:\s*\[\s*"(.+?)"`)
						tokenMatch := r.FindAllStringSubmatch(javascriptText, -1)

						if tokenMatch != nil {
							return tokenMatch[0][1]
						}
					}
				}
			}
		}
	}
}

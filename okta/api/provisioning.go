package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type OktaWebClient struct {
	HostURL  string
	AdminURL string
	UserName string
	Password string
	OrgID    string
}

type OktaAuthResponse struct {
	ExpiresAt    time.Time `json:"expiresAt"`
	SessionToken string    `json:"sessionToken"`
	StateToken   string    `json:"stateToken"`
	Status       string    `json:"status"`
}

func doRequest(client http.Client, request *http.Request) (*http.Response, error) {
	resp, err := client.Do(request)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp, fmt.Errorf("Provisioning did not yield successful %d", resp.StatusCode)
	}

	return resp, nil
}

func (o *OktaWebClient) configureAWSProvisioning(appID string, accessKey string, secretKey string) error {
	client := http.Client{}
	log.Println("[DEBUG] Running AWS provisioning method...")
	authBody := fmt.Sprintf(`{"username":"%s", "password":"%s"}`, o.UserName, o.Password)

	cookieJar, _ := cookiejar.New(nil)
	client.Jar = cookieJar

	authUrl := fmt.Sprintf(`%s/api/v1/authn`, o.HostURL)
	req, _ := http.NewRequest("POST", authUrl, strings.NewReader(authBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := doRequest(client, req)
	if err != nil {
		log.Println("[ERROR] AWS provisioning: Failed to POST to authn route....")
		log.Println(authBody)
		return err
	}

	defer res.Body.Close()

	oktaAuthResponse := &OktaAuthResponse{}
	body, err := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &oktaAuthResponse)

	cookieUrl := fmt.Sprintf("%s/login/sessionCookieRedirect?checkAccountSetupComplete=true&token=%s&redirectUrl=%s/user/notifications", o.HostURL, oktaAuthResponse.SessionToken, o.HostURL)

	req2, _ := http.NewRequest("GET", cookieUrl, nil)
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Accept", "application/json")

	_, err = doRequest(client, req2)
	if err != nil {
		log.Println("[ERROR] AWS provisioning: Failed to GET to sessionCookieRedirect route....")
		log.Println(cookieUrl)
		return err
	}

	// ---------------
	userHomeUrl := fmt.Sprintf("%s/app/UserHome", o.HostURL)
	req3, _ := http.NewRequest("GET", userHomeUrl, nil)
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Accept", "application/json")

	_, err = doRequest(client, req3)
	if err != nil {
		log.Println("[ERROR] AWS provisioning: Failed to GET to userHomeUrl route....")
		log.Println(userHomeUrl)
		return err
	}

	// ---------------
	adminEntryUrl := fmt.Sprintf("%s/home/admin-entry", o.HostURL)
	req4, _ := http.NewRequest("GET", adminEntryUrl, nil)
	req4.Header.Set("Content-Type", "application/json")
	req4.Header.Set("Accept", "application/json")

	_, err = doRequest(client, req4)
	if err != nil {
		log.Println("[ERROR] AWS provisioning: Failed to GET to admin-entry route....")
		log.Println(adminEntryUrl)
		return err
	}

	// ---------------
	adminSsoUrl := fmt.Sprintf("%s/admin/sso/oidc-entry", o.AdminURL)
	req5, _ := http.NewRequest("GET", adminSsoUrl, nil)
	req5.Header.Set("Content-Type", "application/json")
	req5.Header.Set("Accept", "application/json")

	_, err = doRequest(client, req5)
	if err != nil {
		log.Println("[ERROR] AWS provisioning: Failed to GET to admin sso oidc-entry route....")
		log.Println(adminSsoUrl)
		return err
	}

	// ---------------
	dashboardUrl := fmt.Sprintf("%s/admin/dashboard", o.AdminURL)
	req6, _ := http.NewRequest("GET", dashboardUrl, nil)
	req6.Header.Set("Content-Type", "application/json")
	req6.Header.Set("Accept", "application/json")

	dashResp, err := doRequest(client, req6)
	if err != nil {
		log.Println("[ERROR] AWS provisioning: Failed to GET to admin dashboard route....")
		log.Println(dashboardUrl)
		return err
	}

	xsrfToken := getXsrfToken(dashResp.Body)
	defer dashResp.Body.Close()

	// ---------------
	appUpdateUrl := fmt.Sprintf("%s/admin/app/amazon_aws/instance/%s/settings/user-mgmt", o.AdminURL, appID)
	updateAppData := url.Values{}
	updateAppData.Add("_xsrfToken", xsrfToken)
	updateAppData.Add("_enabled", "on")
	updateAppData.Add("accessKeyUM", accessKey)
	updateAppData.Add("secretKeyUM", secretKey)
	updateAppData.Add("accountIds", "")
	updateAppData.Add("_pushNewAccount", "on")
	updateAppData.Add("_pushProfile", "on")
	updateAppData.Add("overrideApiURL", "")
	updateAppData.Add("pushNewAccount", "true")

	if accessKey == "" || secretKey == "" {
		updateAppData.Add("enabled", "false")
	} else {
		updateAppData.Add("enabled", "true")
	}

	req7, _ := http.NewRequest("POST", appUpdateUrl, strings.NewReader(updateAppData.Encode()))
	req7.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req7.Header.Set("Accept", "application/json")

	//here we are not successfully updating it
	_, err = doRequest(client, req7)
	if err != nil {
		log.Println("[ERROR] AWS provisioning: Failed to POST to app update route....")
		log.Println(updateAppData.Encode())
		return err
	}

	log.Println("[DEBUG] Successfully ran AWS provisioning method...")
	return nil
}

func (o *OktaWebClient) RevokeAWSProvisioning(appID string) error {
	return o.configureAWSProvisioning(appID, "", "")
}

func (o *OktaWebClient) SetAWSProvisioning(appID string, accessKey string, secretKey string) error {
	return o.configureAWSProvisioning(appID, accessKey, secretKey)
}

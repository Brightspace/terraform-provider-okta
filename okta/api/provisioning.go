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
		return resp, fmt.Errorf("Provisioning did not yield successful %s", resp.StatusCode)
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
	oneUrl := fmt.Sprintf("%s/home/saasure/%s", o.HostURL, o.OrgID)
	req4, _ := http.NewRequest("GET", oneUrl, nil)
	req4.Header.Set("Content-Type", "application/json")
	req4.Header.Set("Accept", "application/json")

	oneResp, err := doRequest(client, req4)
	if err != nil {
		log.Println("[ERROR] AWS provisioning: Failed to GET to saasure route....")
		log.Println(oneUrl)
		return err
	}

	ssoToken := getSsoToken(oneResp.Body)
	defer oneResp.Body.Close()

	// ---------------
	adminSsoUrl := fmt.Sprintf("%s/admin/sso/request", o.AdminURL)
	postData := url.Values{}
	postData.Add("token", ssoToken)
	req5, _ := http.NewRequest("POST", adminSsoUrl, strings.NewReader(postData.Encode()))
	req5.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req5.Header.Set("Accept", "application/json")

	ssoResp, err := doRequest(client, req5)
	if err != nil {
		log.Println("[ERROR] AWS provisioning: Failed to POST to admin sso route....")
		log.Println(postData.Encode())
		return err
	}

	xsrfToken := getXsrfToken(ssoResp.Body)
	defer ssoResp.Body.Close()

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

	req6, _ := http.NewRequest("POST", appUpdateUrl, strings.NewReader(updateAppData.Encode()))
	req6.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req6.Header.Set("Accept", "application/json")

	//here we are not successfully updating it
	_, err = doRequest(client, req6)
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

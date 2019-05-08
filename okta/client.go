package okta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/matryer/try"
	"golang.org/x/net/html"
)

type Config struct {
	OktaURL      string
	OktaAdminUrl string
	APIKey       string
	UserName     string
	Password     string
	OrgID        string
	RetryMaximum int
}

type OktaClient struct {
	OktaURL      string
	OktaAdminUrl string
	APIKey       string
	UserName     string
	Password     string
	OrgID        string
	RetryMaximum int
}

type OktaAuthResponse struct {
	ExpiresAt    time.Time `json:"expiresAt"`
	SessionToken string    `json:"sessionToken"`
	StateToken   string    `json:"stateToken"`
	Status       string    `json:"status"`
}

type OktaGroup struct {
	ID                    string           `json:"id,omitempty"`
	Created               *time.Time       `json:"created,omitempty"`
	LastUpdated           *time.Time       `json:"lastUpdated,omitempty"`
	LastMembershipUpdated *time.Time       `json:"lastMembershipUpdated,omitempty"`
	ObjectClass           []string         `json:"objectClass,omitempty"`
	Type                  string           `json:"type,omitempty"`
	Profile               OktaGroupProfile `json:"profile"`
}

type OktaGroupProfile struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type OktaGroupSaml struct {
	ID      string               `json:"id,omitempty"`
	Profile OktaGroupProfileSaml `json:"profile"`
}

type OktaAppUser struct {
	ID      string               `json:"id,omitempty"`
	Scope   string               `json:"scope,omitempty"`
	Profile OktaGroupProfileSaml `json:"profile"`
}

type OktaGroupProfileSaml struct {
	Role      string   `json:"role,omitempty"`
	SamlRoles []string `json:"samlRoles,omitempty"`
}

type OktaUser struct {
	ID              string          `json:"id"`
	Status          string          `json:"status"`
	Created         *time.Time      `json:"created,omitempty"`
	Activated       *time.Time      `json:"activated,omitempty"`
	StatusChanged   *time.Time      `json:"statusChanged,omitempty"`
	LastLogin       *time.Time      `json:"lastLogin,omitempty"`
	LastUpdated     *time.Time      `json:"lastUpdated,omitempty"`
	PasswordChanged *time.Time      `json:"passwordChanged,omitempty"`
	Profile         OktaUserProfile `json:"profile,omitempty"`
}

type OktaUserProfile struct {
	Login       string   `json:"login,omitempty"`
	FirstName   string   `json:"firstName,omitempty"`
	LastName    string   `json:"lastName,omitempty"`
	NickName    string   `json:"nickName,omitempty"`
	DisplayName string   `json:"displayName,omitempty"`
	Email       string   `json:"email,omitempty"`
	SecondEmail string   `json:"secondEmail,omitempty"`
	Role        string   `json:"role,omitempty"`
	SamlRoles   []string `json:"samlRoles,omitempty"`
}

func NewClient(c *Config) OktaClient {
	client = initHTTPClient()

	return OktaClient{
		OktaURL:      c.OktaURL,
		OktaAdminUrl: c.OktaAdminUrl,
		APIKey:       c.APIKey,
		UserName:     c.UserName,
		Password:     c.Password,
		OrgID:        c.OrgID,
		RetryMaximum: c.RetryMaximum,
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

func (o *OktaClient) SendRequest(url string, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	err := try.Do(func(ampt int) (bool, error) {
		var err error
		resp, err = client.Do(req)
		if err != nil {
			log.Printf("[DEBUG] (%d) retrying request: (Attempt: %d/%d, URL: %q)", resp.StatusCode, ampt, o.RetryMaximum, err)
			time.Sleep(30 * time.Second)
		} else if resp.StatusCode == 200 || resp.StatusCode == 204 {
			return false, nil
		} else if resp.StatusCode == 429 {
			log.Printf("[DEBUG] Rate limit hit (%d) retrying request: (Attempt: %d/%d, URL: %s)", resp.StatusCode, ampt, o.RetryMaximum, url)
			time.Sleep(45 * time.Second)
		} else if resp.StatusCode == 404 {
			log.Printf("[DEBUG] Resource not found (%d): (Attempt: %d/%d, URL: %s)", resp.StatusCode, ampt, o.RetryMaximum, url)
			return false, nil
		} else if resp.StatusCode != 200 {
			log.Printf("[DEBUG] bad status code (%d) retrying request: (Attempt: %d/%d, URL: %s)", resp.StatusCode, ampt, o.RetryMaximum, url)
			time.Sleep(30 * time.Second)
		}

		retry := ampt < o.RetryMaximum
		if !retry && resp.StatusCode == 429 {
			return retry, fmt.Errorf("Rate limit prevented the completion of the request: %s", url)
		}

		return retry, err
	})
	return resp, err
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

	resp, err := o.SendRequest(url, req)
	if err != nil {
		return idApp, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		msg := buf.String()

		return idApp, fmt.Errorf("Error creating application in Okta: %s", msg)
	}

	err = json.NewDecoder(resp.Body).Decode(&idApp)
	if err != nil {
		return idApp, err
	}

	return idApp, nil
}

func (o *OktaClient) UpdateApplication(application Application) (Application, error) {
	var app Application
	url := fmt.Sprintf("%s/api/v1/apps/%s", o.OktaURL, application.ID)

	body, err := json.Marshal(application)
	if err != nil {
		return app, err
	}

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	resp, err := o.SendRequest(url, req)
	if err != nil {
		return app, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		msg := buf.String()

		return app, fmt.Errorf("Error updating application in Okta: %s \nRequest Body: %s", msg, string(body))
	}

	err = json.NewDecoder(resp.Body).Decode(&app)
	if err != nil {
		return app, err
	}

	return app, nil
}

func (o *OktaClient) ReadApplication(appID string) (IdentifiedApplication, bool, error) {
	var app IdentifiedApplication
	url := fmt.Sprintf("%s/api/v1/apps/%s", o.OktaURL, appID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	resp, err := o.SendRequest(url, req)
	if err != nil {
		return app, false, err
	}

	if resp.StatusCode == 404 {
		return app, true, nil
	}

	err = json.NewDecoder(resp.Body).Decode(&app)
	if err != nil {
		return app, false, err
	}

	return app, false, err
}

func (o *OktaClient) GetSAMLMetaData(appID string, keyID string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/apps/%s/sso/saml/metadata?kid=%s", o.OktaURL, appID, keyID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	rate_guard := 10
	for rate_guard > 0 {
		resp, err := o.SendRequest(url, req)
		if err != nil {
			return "", err
		}
	
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		samlMetaData := buf.String()

		// WORKAROUND:
		// Rate limit workaround: Returns rate limit error in SAML
		// rather than as status code
		if (strings.Contains(samlMetaData, "E0000047")) {
			rate_guard = rate_guard - 1
			continue
		}

		rate_guard = -1
		return samlMetaData, nil
	}

	return "", nil	
}

func (o *OktaClient) DeleteApplication(appID string) error {
	// Deactivate app first
	url := fmt.Sprintf("%s/api/v1/apps/%s/lifecycle/deactivate", o.OktaURL, appID)

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	res, err := o.SendRequest(url, req)
	if err != nil {
		return err
	}

	// Delete app
	url = fmt.Sprintf("%s/api/v1/apps/%s", o.OktaURL, appID)

	req, _ = http.NewRequest("DELETE", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	res, err = o.SendRequest(url, req)
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

func (o *OktaClient) RemoveMemberFromApp(appId string, userId string) error {
	url := fmt.Sprintf("%s/api/v1/apps/%s/users/%s", o.OktaURL, appId, userId)

	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	_, err := o.SendRequest(url, req)
	if err != nil {
		return err
	}

	return nil
}

func (o *OktaClient) GetAppMember(appId string, userId string) (OktaUser, error) {
	var user OktaUser
	url := fmt.Sprintf("%s/api/v1/apps/%s/users/%s", o.OktaURL, appId, userId)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	res, err := o.SendRequest(url, req)
	if err != nil {
		return user, err
	}

	if res.StatusCode == 404 {
		return user, nil
	}

	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (o *OktaClient) ListAppMembers(appId string) ([]OktaUser, error) {
	oktaUsers := make([]OktaUser, 0)
	url := fmt.Sprintf("%s/api/v1/apps/%s/users", o.OktaURL, appId)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	res, err := o.SendRequest(url, req)
	if err != nil {
		return oktaUsers, err
	}

	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&oktaUsers)
	if err != nil {
		return oktaUsers, err
	}

	return oktaUsers, nil
}

func (o *OktaClient) AddMemberToApp(appId string, userId string, role string, roles []string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/apps/%s/users", o.OktaURL, appId)

	var memberInput OktaAppUser
	memberInput.ID = userId
	memberInput.Scope = "USER"
	memberInput.Profile.Role = role
	memberInput.Profile.SamlRoles = roles

	body, err := json.Marshal(memberInput)
	if err != nil {
		return "", err
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	_, err = o.SendRequest(url, req)
	if err != nil {
		return "", err
	}

	return "", nil
}

func (o *OktaClient) GetUserIDByEmail(user string) (string, error) {
	var oktaUser []OktaUser
	url := fmt.Sprintf("%s/api/v1/users?q=%s", o.OktaURL, user)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	resp, err := o.SendRequest(url, req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&oktaUser)
	if err != nil {
		return "", err
	}

	for _, user := range oktaUser {
		if strings.Contains(user.Profile.Login, "desire2learn.com") {
			return user.ID, nil
		}
	}

	return "", fmt.Errorf("Could not find user in desire2learn domain for email %s", user)
}

func (o *OktaClient) DelayRateLimit(appID string) error {
	url := fmt.Sprintf("%s/api/v1/apps/%s", o.OktaURL, appID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	err := try.Do(func(ampt int) (bool, error) {
		resp, err := client.Do(req)
		retry := ampt < o.RetryMaximum

		if err != nil || resp.StatusCode != 200 {
			log.Printf("[DEBUG] (%d) retrying request: (Attempt: %d/%d, URL: %q)", resp.StatusCode, ampt, o.RetryMaximum, err)
			time.Sleep(30 * time.Second)
		} else if resp.StatusCode == 429 {
			log.Printf("[DEBUG] Rate limit hit (%d) retrying request: (Attempt: %d/%d, URL: %s)", resp.StatusCode, ampt, o.RetryMaximum, url)
			time.Sleep(45 * time.Second)
		} else if resp.StatusCode != 200 {
			log.Printf("[DEBUG] bad status code (%d) retrying request: (Attempt: %d/%d, URL: %s)", resp.StatusCode, ampt, o.RetryMaximum, url)
			time.Sleep(30 * time.Second)
		}

		limit, err := strconv.Atoi(resp.Header.Get("X-Rate-Limit-Limit"))
		if err != nil {
			return retry, err
		}

		remaining, err := strconv.Atoi(resp.Header.Get("X-Rate-Limit-Remaining"))
		if err != nil {
			return retry, err
		}

		ratio := (remaining * 100) / limit

		if ratio < 50 {
			log.Printf("[DEBUG] remaining retries to low, retrying request: (Attempt: %d/%d, URL: %s)", resp.StatusCode, ampt, o.RetryMaximum, url)
			time.Sleep(55 * time.Second)
		}

		if !retry && resp.StatusCode == 429 {
			return retry, fmt.Errorf("Rate limit prevented the completion of the request: %s", url)
		}

		return retry, err
	})
	if err != nil {
		return err
	}

	return nil
}

func (o *OktaClient) RevokeProvisioningSettings(appID string) error {
	log.Println("[DEBUG] Running RevokeProvisioningSettings method...")
	authBody := fmt.Sprintf(`{"username":"%s", "password":"%s"}`, o.UserName, o.Password)

	cookieJar, _ := cookiejar.New(nil)
	client.Jar = cookieJar

	err := o.DelayRateLimit(appID)
	if err != nil {
		return err
	}

	authUrl := fmt.Sprintf(`%s/api/v1/authn`, o.OktaURL)
	req, _ := http.NewRequest("POST", authUrl, strings.NewReader(authBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Println("[ERROR] RevokeProvisioningSettings: Failed to POST to authn route....")
		log.Println(authBody)
		panic(err)
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
		log.Println("[ERROR] RevokeProvisioningSettings: Failed to GET to sessionCookieRedirect route....")
		log.Println(cookieUrl)
		panic(err2)
	}

	// ---------------
	userHomeUrl := fmt.Sprintf("%s/app/UserHome", o.OktaURL)
	req3, _ := http.NewRequest("GET", userHomeUrl, nil)
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Accept", "application/json")

	_, err3 := client.Do(req3)
	if err3 != nil {
		log.Println("[ERROR] RevokeProvisioningSettings: Failed to GET to userHomeUrl route....")
		log.Println(userHomeUrl)
		panic(err3)
	}

	// ---------------
	oneUrl := fmt.Sprintf("%s/home/saasure/%s", o.OktaURL, o.OrgID)
	req4, _ := http.NewRequest("GET", oneUrl, nil)
	req4.Header.Set("Content-Type", "application/json")
	req4.Header.Set("Accept", "application/json")

	oneResp, err4 := client.Do(req4)
	if err4 != nil {
		log.Println("[ERROR] RevokeProvisioningSettings: Failed to GET to saasure route....")
		log.Println(oneUrl)
		panic(err4)
	}

	ssoToken := getSsoToken(oneResp.Body)
	defer oneResp.Body.Close()

	// ---------------
	adminSsoUrl := fmt.Sprintf("%s/admin/sso/request", o.OktaAdminUrl)
	postData := url.Values{}
	postData.Add("token", ssoToken)
	req5, _ := http.NewRequest("POST", adminSsoUrl, strings.NewReader(postData.Encode()))
	req5.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req5.Header.Set("Accept", "application/json")

	ssoResp, err5 := client.Do(req5)
	if err5 != nil {
		log.Println("[ERROR] RevokeProvisioningSettings: Failed to POST to admin sso route....")
		log.Println(postData.Encode())
		panic(err5)
	}

	xsrfToken := getXsrfToken(ssoResp.Body)
	defer ssoResp.Body.Close()

	// ---------------
	appUpdateUrl := fmt.Sprintf("%s/admin/app/amazon_aws/instance/%s/settings/user-mgmt", o.OktaAdminUrl, appID)
	updateAppData := url.Values{}
	updateAppData.Add("_xsrfToken", xsrfToken)
	updateAppData.Add("enabled", "false")
	updateAppData.Add("_enabled", "on")
	updateAppData.Add("accessKeyUM", "")
	updateAppData.Add("secretKeyUM", "")
	updateAppData.Add("accountIds", "")
	updateAppData.Add("_pushNewAccount", "on")
	updateAppData.Add("_pushProfile", "on")
	updateAppData.Add("overrideApiURL", "")
	updateAppData.Add("pushNewAccount", "true")

	//inputs are correct

	req6, _ := http.NewRequest("POST", appUpdateUrl, strings.NewReader(updateAppData.Encode()))
	req6.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req6.Header.Set("Accept", "application/json")

	//here we are not successfully updating it
	_, err6 := client.Do(req6)
	if err6 != nil {
		log.Println("[ERROR] RevokeProvisioningSettings: Failed to POST to app update route....")
		log.Println(updateAppData.Encode())
		panic(err6)
	}

	log.Println("[DEBUG] Successfully ran RevokeProvisioningSettings method...")
	return nil
}

func (o *OktaClient) SetProvisioningSettings(appID string, oktaAWSKey string, oktaAWSSecretKey string) error {
	log.Println("[DEBUG] Running SetProvisioningSettings method...")
	authBody := fmt.Sprintf(`{"username":"%s", "password":"%s"}`, o.UserName, o.Password)

	cookieJar, _ := cookiejar.New(nil)
	client.Jar = cookieJar

	err := o.DelayRateLimit(appID)
	if err != nil {
		return err
	}

	authUrl := fmt.Sprintf(`%s/api/v1/authn`, o.OktaURL)
	req, _ := http.NewRequest("POST", authUrl, strings.NewReader(authBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Println("[ERROR] SetProvisioningSettings: Failed to POST to authn route....")
		log.Println(authBody)
		panic(err)
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
		log.Println("[ERROR] SetProvisioningSettings: Failed to GET to sessionCookieRedirect route....")
		log.Println(cookieUrl)
		panic(err2)
	}

	// ---------------
	userHomeUrl := fmt.Sprintf("%s/app/UserHome", o.OktaURL)
	req3, _ := http.NewRequest("GET", userHomeUrl, nil)
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Accept", "application/json")

	_, err3 := client.Do(req3)
	if err3 != nil {
		log.Println("[ERROR] SetProvisioningSettings: Failed to GET to userHomeUrl route....")
		log.Println(userHomeUrl)
		panic(err3)
	}

	// ---------------
	oneUrl := fmt.Sprintf("%s/home/saasure/%s", o.OktaURL, o.OrgID)
	req4, _ := http.NewRequest("GET", oneUrl, nil)
	req4.Header.Set("Content-Type", "application/json")
	req4.Header.Set("Accept", "application/json")

	oneResp, err4 := client.Do(req4)
	if err4 != nil {
		log.Println("[ERROR] SetProvisioningSettings: Failed to GET to saasure route....")
		log.Println(oneUrl)
		panic(err4)
	}

	ssoToken := getSsoToken(oneResp.Body)
	defer oneResp.Body.Close()

	// ---------------
	adminSsoUrl := fmt.Sprintf("%s/admin/sso/request", o.OktaAdminUrl)
	postData := url.Values{}
	postData.Add("token", ssoToken)
	req5, _ := http.NewRequest("POST", adminSsoUrl, strings.NewReader(postData.Encode()))
	req5.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req5.Header.Set("Accept", "application/json")

	ssoResp, err5 := client.Do(req5)
	if err5 != nil {
		log.Println("[ERROR] SetProvisioningSettings: Failed to POST to admin sso route....")
		log.Println(postData.Encode())
		panic(err5)
	}

	xsrfToken := getXsrfToken(ssoResp.Body)
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

	//inputs are correct

	req6, _ := http.NewRequest("POST", appUpdateUrl, strings.NewReader(updateAppData.Encode()))
	req6.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req6.Header.Set("Accept", "application/json")

	//here we are not successfully updating it
	_, err6 := client.Do(req6)
	if err6 != nil {
		log.Println("[ERROR] SetProvisioningSettings: Failed to POST to app update route....")
		log.Println(updateAppData.Encode())
		panic(err6)
	}

	log.Println("[DEBUG] Successfully ran SetProvisioningSettings method...")
	return nil
}

func getXsrfToken(resBody io.Reader) string {
	z := html.NewTokenizer(resBody)

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

func getSsoToken(resBody io.Reader) string {
	z := html.NewTokenizer(resBody)

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

package api

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

	"github.com/go-resty/resty/v2"
	"github.com/matryer/try"
	"golang.org/x/net/html"
)

const MaximumRetryWaitTimeInSeconds = 15 * time.Minute
const RetryWaitTimeInSeconds = 30 * time.Second

type Settings struct {
	App AppSettings `json:"app"`
}

type AppSettings struct {
	AwsEnvironmentType  string `json:"awsEnvironmentType"`
	GroupFilter         string `json:"groupFilter"`
	LoginURL            string `json:"loginUrl"`
	JoinAllRoles        bool   `json:"joinAllRoles"`
	SessionDuration     int    `json:"sessionDuration"`
	RoleValuePattern    string `json:"roleValuePattern"`
	IdentityProviderArn string `json:"identityProviderArn"`
	AccessKey           string `json:"accessKey"`
	SecretKey           string `json:"secretKey"`
}

type Application struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Label      string   `json:"label"`
	Features   []string `json:"features"`
	SignOnMode string   `json:"signOnMode"`
	Settings   Settings `json:"settings"`
}

type IdentifiedApplication struct {
	Application
	Credentials Credentials `json:"credentials"`
}

type Credentials struct {
	Signing Signing `json:"signing"`
}

type Signing struct {
	KeyID string `json:"kid"`
}

////
////
////
type Okta struct {
	APIKey       string
	HostURL      string
	OrgID        string
	RetryMaximum int
	RestClient   *resty.Client
}

func (o *Okta) GetApplication(appID string) (*IdentifiedApplication, error) {
	restClient := o.GetRestClient()

	url := fmt.Sprintf("/api/v1/apps/%s", appID)
	req := restClient.R().SetBody("").SetResult(&IdentifiedApplication{})

	resp, err := req.Get(url)
	if err != nil {
		return nil, err
	}

	status := resp.StatusCode()
	if status == http.StatusNotFound {
		return nil, nil
	}

	response := resp.Result().(*IdentifiedApplication)
	if response == nil {
		return nil, nil
	}

	return response, nil
}

func (o *Okta) CreateApplication(application Application) (*IdentifiedApplication, error) {
	var result *IdentifiedApplication
	restClient := o.GetRestClient()

	body, err := json.Marshal(application)
	if err != nil {
		return result, err
	}

	url := "/api/v1/apps"
	req := restClient.R().SetBody(string(body)).SetResult(&IdentifiedApplication{})

	resp, err := req.Post(url)
	if err != nil {
		return result, err
	}

	response := resp.Result().(*IdentifiedApplication)
	return response, nil
}

func (o *Okta) DeactivateApplication(appID string) error {
	restClient := cloudability.GetRestClient()

	url := fmt.Sprintf("/api/v1/apps/%s/lifecycle/deactivate", appID)
	req := restClient.R().SetBody("")

	_, err := req.Post(url)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (o *Okta) DeleteApplication(appID string) error {
	restClient := cloudability.GetRestClient()

	url := fmt.Sprintf("%s/api/v1/apps/%s", appID)
	req := restClient.R().SetBody("")

	_, err := req.Delete(url)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (o *Okta) GetSAMLMetadata(appID string, keyID string) (string, error) {
	restClient := o.GetRestClient()

	url := fmt.Sprintf("/api/v1/apps/%s/sso/saml/metadata?kid=%s", appID, keyID)
	req := restClient.R().SetBody("")
	req.SetHeader("Accept", "application/xml")

	resp, err := req.Get(url)
	if err != nil {
		return "", err
	}

	status := resp.StatusCode()
	if status == http.StatusNotFound {
		return "", nil
	}

	response := string(resp.Body())
	return response, nil
}

func (o *Okta) UpdateApplication(application Application) (*Application, error) {
	var result *IdentifiedApplication
	restClient := o.GetRestClient()

	body, err := json.Marshal(application)
	if err != nil {
		return result, err
	}

	url := fmt.Sprintf("/api/v1/apps/%s", application.ID)
	req := restClient.R().SetBody(string(body)).SetResult(&IdentifiedApplication{})

	resp, err := req.Put(url)
	if err != nil {
		return result, err
	}

	response := resp.Result().(*IdentifiedApplication)
	return response, nil
}

func (okta *Okta) SetRestClient(rest *resty.Client) {
	rest.SetHostURL(okta.HostURL)

	// Retry
	rest.SetRetryCount(okta.RetryMaximum)
	rest.SetRetryWaitTime(RetryWaitTimeInSeconds)
	rest.SetRetryMaxWaitTime(MaximumRetryWaitTimeInSeconds)
	rest.AddRetryCondition(func(r *resty.Response, err error) bool {
		switch code := r.StatusCode(); code {
		case http.StatusTooManyRequests:
			return true
		default:
			return false
		}
	})

	// Error handling
	rest.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		status := r.StatusCode()
		if status == http.StatusNotFound {
			return nil
		}

		if (status < 200) || (status >= 400) {
			return fmt.Errorf("Response not successful: Received status code %d.", status)
		}

		return nil
	})

	sign, _ := NewHTTPSignature(okta.APIKey)
	rest.SetHeaders(sign)

	okta.RestClient = rest
}

func (okta *Okta) GetRestClient() *resty.Client {
	if okta.RestClient == nil {
		rest := resty.New()
		okta.SetRestClient(rest)
	}
	return okta.RestClient
}

func (o *Okta) GetUserIDByEmail(user string) (string, error) {

	restClient := o.GetRestClient()
	url := fmt.Sprintf("/api/v1/users?q=%s", user)

	req := restClient.R().SetBody("").SetResult(&[]OktaUser{})

	resp, err := req.Get(url)
	if err != nil {
		return "", err
	}

	status := resp.StatusCode()
	if status == http.StatusNotFound {
		return "", nil
	}

	result, err := resp.Result().([]*OktaUser)
	for _, user := range result {
		if strings.Contains(user.Profile.Login, "desire2learn.com") {
			return user.ID, nil
		}
	}

	return "", nil
}

///
///
///
///

type OktaClient struct {
	OktaURL      string
	OktaAdminUrl string
	APIKey       string
	UserName     string
	Password     string
	OrgID        string
	RetryMaximum int
	RestClient   http.Client
}

type OktaAuthResponse struct {
	ExpiresAt    time.Time `json:"expiresAt"`
	SessionToken string    `json:"sessionToken"`
	StateToken   string    `json:"stateToken"`
	Status       string    `json:"status"`
}

type OktaAppUser struct {
	ID      string               `json:"id,omitempty"`
	Scope   string               `json:"scope,omitempty"`
	Profile OktaGroupProfileSaml `json:"profile"`
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

func (o *OktaClient) SendRequest(url string, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	client := o.RestClient
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
	resultsPerPage := 500
	url := fmt.Sprintf("%s/api/v1/apps/%s/users?limit=%s", o.OktaURL, appId, resultsPerPage)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("SSWS %s", o.APIKey))

	var resp *http.Response
	err := try.Do(func(ampt int) (bool, error) {
		var err error
		retry := ampt < o.RetryMaximum

		resp, err = o.SendRequest(url, req)
		if err != nil {
			return retry, err
		}

		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&oktaUsers)
		if err != nil {
			return retry, err
		}

		return retry, err
	})
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

func (o *OktaClient) RevokeProvisioningSettings(appID string) error {
	client := o.RestClient
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
	client := o.RestClient
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

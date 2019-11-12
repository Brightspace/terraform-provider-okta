package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const MaximumRetryWaitTimeInSeconds = 15 * time.Minute
const RetryWaitTimeInSeconds = 30 * time.Second

type Application struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Label      string   `json:"label"`
	Features   []string `json:"features"`
	SignOnMode string   `json:"signOnMode"`
	Settings   struct {
		App struct {
			AwsEnvironmentType  string `json:"awsEnvironmentType"`
			GroupFilter         string `json:"groupFilter"`
			LoginURL            string `json:"loginUrl"`
			JoinAllRoles        bool   `json:"joinAllRoles"`
			SessionDuration     int    `json:"sessionDuration"`
			RoleValuePattern    string `json:"roleValuePattern"`
			IdentityProviderArn string `json:"identityProviderArn"`
			AccessKey           string `json:"accessKey"`
			SecretKey           string `json:"secretKey"`
		} `json:"app"`
	} `json:"settings"`
}

type IdentifiedApplication struct {
	Application
	Credentials struct {
		Signing struct {
			KeyID string `json:"kid"`
		} `json:"signing"`
	} `json:"credentials"`
}

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
	restClient := o.GetRestClient()

	url := fmt.Sprintf("/api/v1/apps/%s/lifecycle/deactivate", appID)
	req := restClient.R().SetBody("")

	_, err := req.Post(url)
	if err != nil {
		return err
	}

	return nil
}

func (o *Okta) DeleteApplication(appID string) error {
	restClient := o.GetRestClient()

	url := fmt.Sprintf("%s/api/v1/apps/%s", appID)
	req := restClient.R().SetBody("")

	_, err := req.Delete(url)
	if err != nil {
		return err
	}

	return nil
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
	var result *Application
	restClient := o.GetRestClient()

	body, err := json.Marshal(application)
	if err != nil {
		return result, err
	}

	url := fmt.Sprintf("/api/v1/apps/%s", application.ID)
	req := restClient.R().SetBody(string(body)).SetResult(&Application{})

	resp, err := req.Put(url)
	if err != nil {
		return result, err
	}

	response := resp.Result().(*Application)
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

	result := resp.Result().([]OktaUser)
	for _, user := range result {
		if strings.Contains(user.Profile.Login, "desire2learn.com") {
			return user.ID, nil
		}
	}

	return "", nil
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

func (o *Okta) RemoveAppMember(appId string, userId string) error {
	restClient := o.GetRestClient()

	url := fmt.Sprintf("/api/v1/apps/%s/users/%s", appId, userId)
	req := restClient.R().SetBody("")

	_, err := req.Delete(url)
	if err != nil {
		return err
	}

	return nil
}

func (o *Okta) GetAppMember(appId string, userId string) (*OktaUser, error) {
	restClient := o.GetRestClient()

	url := fmt.Sprintf("/api/v1/apps/%s/users/%s", appId, userId)
	req := restClient.R().SetBody("").SetResult(&OktaUser{})

	resp, err := req.Get(url)
	if err != nil {
		return nil, err
	}

	status := resp.StatusCode()
	if status == http.StatusNotFound {
		return nil, nil
	}

	response := resp.Result().(*OktaUser)
	if response == nil {
		return nil, nil
	}

	return response, nil
}

func (o *Okta) ListAppMembers(appId string) ([]OktaUser, error) {
	resultsPerPage := 500
	restClient := o.GetRestClient()

	url := fmt.Sprintf("%s/api/v1/apps/%s/users?limit=%s", appId, resultsPerPage)
	req := restClient.R().SetBody("").SetResult(&[]OktaUser{})

	resp, err := req.Get(url)
	if err != nil {
		return nil, err
	}

	status := resp.StatusCode()
	if status == http.StatusNotFound {
		return nil, nil
	}

	response := resp.Result().([]OktaUser)
	if response == nil {
		return nil, nil
	}

	return response, nil
}

func (o *Okta) AddAppMember(appId string, userId string, role string, roles []string) (string, error) {
	restClient := o.GetRestClient()

	url := fmt.Sprintf("/api/v1/apps/%s/users", appId)
	payload := map[string]interface{}{
		"id":    userId,
		"scope": "USER",
		"profile": map[string]interface{}{
			"role":      role,
			"samlRoles": roles,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req := restClient.R().SetBody(string(body)).SetResult(&OktaUser{})

	_, err = req.Post(url)
	if err != nil {
		return "", err
	}

	// result := resp.Result().(*OktaUser)
	return "", nil
}

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

type OktaApplicationContents struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Label       string                  `json:"label"`
	SignOnMode  string                  `json:"signOnMode"`
	Settings    OktaApplicationSettings `json:"settings,omitempty"`
}

type OktaApplication struct {
	OktaApplicationContents
	Credentials struct {
		Signing struct {
			KeyID string `json:"kid,omitempty"`
		} `json:"signing,omitempty"`
	} `json:"credentials,omitempty"`
}

type OktaApplicationSettings struct {
	App OktaApplicationAppSettings `json:"app,omitempty"`
}

type OktaApplicationAppSettings struct {
	AwsEnvironmentType  string `json:"awsEnvironmentType,omitempty"`
	GroupFilter         string `json:"groupFilter,omitempty"`
	LoginURL            string `json:"loginUrl,omitempty"`
	JoinAllRoles        bool   `json:"joinAllRoles,omitempty"`
	SessionDuration     int    `json:"sessionDuration,omitempty"`
	RoleValuePattern    string `json:"roleValuePattern,omitempty"`
	IdentityProviderArn string `json:"identityProviderArn,omitempty"`
}

type OktaUser struct {
	ID              string     `json:"id"`
	Status          string     `json:"status"`
	Created         *time.Time `json:"created,omitempty"`
	Activated       *time.Time `json:"activated,omitempty"`
	StatusChanged   *time.Time `json:"statusChanged,omitempty"`
	LastLogin       *time.Time `json:"lastLogin,omitempty"`
	LastUpdated     *time.Time `json:"lastUpdated,omitempty"`
	PasswordChanged *time.Time `json:"passwordChanged,omitempty"`
	Profile         struct {
		Login       string   `json:"login,omitempty"`
		FirstName   string   `json:"firstName,omitempty"`
		LastName    string   `json:"lastName,omitempty"`
		NickName    string   `json:"nickName,omitempty"`
		DisplayName string   `json:"displayName,omitempty"`
		Email       string   `json:"email,omitempty"`
		SecondEmail string   `json:"secondEmail,omitempty"`
		Role        string   `json:"role,omitempty"`
		SamlRoles   []string `json:"samlRoles,omitempty"`
	} `json:"profile,omitempty"`
}

type Okta struct {
	APIKey       string
	HostURL      string
	OrgID        string
	RetryMaximum int
	RestClient   *resty.Client
}

func (o *Okta) GetApplication(appID string) (*OktaApplication, error) {
	restClient := o.GetRestClient()

	url := fmt.Sprintf("/api/v1/apps/%s", appID)
	req := restClient.R().SetBody("").SetResult(&OktaApplication{})

	resp, err := req.Get(url)
	if err != nil {
		return nil, err
	}

	status := resp.StatusCode()
	if status == http.StatusNotFound {
		return nil, nil
	}

	response := resp.Result().(*OktaApplication)
	if response == nil {
		return nil, nil
	}

	return response, nil
}

func (o *Okta) CreateAwsApplication(name string, providerArn string) (*OktaApplication, error) {
	application := OktaApplicationContents{
		Name:       "amazon_aws",
		Label:      name,
		SignOnMode: "SAML_2_0",
		Settings: OktaApplicationSettings{
			App: OktaApplicationAppSettings{
				AwsEnvironmentType:  "aws.amazon",
				LoginURL:            "https://console.aws.amazon.com/ec2/home",
				JoinAllRoles:        false,
				SessionDuration:     43200,
				IdentityProviderArn: providerArn,
			},
		},
	}

	return o.CreateApplication(application)
}

func (o *Okta) CreateApplication(application OktaApplicationContents) (*OktaApplication, error) {
	var result *OktaApplication
	restClient := o.GetRestClient()

	body, err := json.Marshal(application)
	if err != nil {
		return result, err
	}

	url := "/api/v1/apps"
	req := restClient.R().SetBody(string(body)).SetResult(&OktaApplication{})

	resp, err := req.Post(url)
	if err != nil {
		return result, err
	}

	response := resp.Result().(*OktaApplication)
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

	url := fmt.Sprintf("api/v1/apps/%s", appID)
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

func (o *Okta) UpdateAwsApplication(appId string, name string, providerArn string) (*OktaApplication, error) {
	application := OktaApplicationContents{
		ID:         appId,
		Name:       "amazon_aws",
		Label:      name,
		SignOnMode: "SAML_2_0",
		Settings: OktaApplicationSettings{
			App: OktaApplicationAppSettings{
				AwsEnvironmentType:  "aws.amazon",
				LoginURL:            "https://console.aws.amazon.com/ec2/home",
				JoinAllRoles:        false,
				SessionDuration:     43200,
				IdentityProviderArn: providerArn,
			},
		},
	}

	return o.UpdateApplication(application)
}

func (o *Okta) UpdateApplication(application OktaApplicationContents) (*OktaApplication, error) {
	var result *OktaApplication
	restClient := o.GetRestClient()

	body, err := json.Marshal(application)
	if err != nil {
		return result, err
	}

	url := fmt.Sprintf("/api/v1/apps/%s", application.ID)
	req := restClient.R().SetBody(string(body)).SetResult(&OktaApplication{})

	resp, err := req.Put(url)
	if err != nil {
		return result, err
	}

	response := resp.Result().(*OktaApplication)
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
	// I've set the results per page to 500, but if the organization needs
	// more than that this will need pagination enabled.
	// TODO: Add pagination for listing app members
	resultsPerPage := 500
	restClient := o.GetRestClient()

	url := fmt.Sprintf("/api/v1/apps/%s/users?limit=%s", appId, resultsPerPage)
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

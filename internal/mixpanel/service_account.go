package mixpanel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type ServiceAccount struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	// Only returned on creation.
	Token string `json:"token,omitempty"`
}

type ServiceAccountProjectMember struct {
	Id   int64  `json:"id"`
	Role string `json:"role"`
}

type results[T any] struct {
	Results T `json:"results"`
}

// IsNotFound reports whether err is a 404 from the Mixpanel API.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

func (c *Client) organizationId() (int64, error) {
	organizations, err := c.GetOrganizations()
	if err != nil {
		return 0, err
	}
	if len(organizations) == 0 {
		return 0, fmt.Errorf("no Mixpanel organization found")
	}
	// We only support one organization for now
	return organizations[0].Id, nil
}

func (c *Client) doJSON(method, url string, data any) ([]byte, error) {
	payload := &bytes.Buffer{}
	if data != nil {
		if err := json.NewEncoder(payload).Encode(data); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	return c.doRequest(req)
}

// CreateServiceAccount returns the created service account, including its Token (secret).
func (c *Client) CreateServiceAccount(username, role string) (*ServiceAccount, error) {
	orgId, err := c.organizationId()
	if err != nil {
		return nil, err
	}

	data := map[string]any{"username": username}
	if role != "" {
		data["role"] = role
	}

	body, err := c.doJSON("POST", fmt.Sprintf("%s/api/app/organizations/%d/service-accounts", c.HostURL, orgId), data)
	if err != nil {
		return nil, err
	}

	var response results[ServiceAccount]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return &response.Results, nil
}

func (c *Client) ListServiceAccounts() ([]ServiceAccount, error) {
	orgId, err := c.organizationId()
	if err != nil {
		return nil, err
	}

	body, err := c.doJSON("GET", fmt.Sprintf("%s/api/app/organizations/%d/service-accounts", c.HostURL, orgId), nil)
	if err != nil {
		return nil, err
	}

	var response results[[]ServiceAccount]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return response.Results, nil
}

// GetServiceAccount returns nil if the service account does not exist.
// It goes through the list endpoint, whose response shape is the documented one.
func (c *Client) GetServiceAccount(id int64) (*ServiceAccount, error) {
	serviceAccounts, err := c.ListServiceAccounts()
	if err != nil {
		return nil, err
	}
	for _, serviceAccount := range serviceAccounts {
		if serviceAccount.Id == id {
			return &serviceAccount, nil
		}
	}
	return nil, nil
}

func (c *Client) DeleteServiceAccount(id int64) error {
	orgId, err := c.organizationId()
	if err != nil {
		return err
	}

	_, err = c.doJSON("DELETE", fmt.Sprintf("%s/api/app/organizations/%d/service-accounts/%d", c.HostURL, orgId, id), nil)
	return err
}

func (c *Client) AddServiceAccountToProject(serviceAccountId, projectId int64, role string) error {
	orgId, err := c.organizationId()
	if err != nil {
		return err
	}

	_, err = c.doJSON("POST", fmt.Sprintf("%s/api/app/organizations/%d/service-accounts/add-to-project", c.HostURL, orgId), map[string]any{
		"projects":            []map[string]any{{"id": projectId, "role": role}},
		"service_account_ids": []int64{serviceAccountId},
	})
	return err
}

func (c *Client) RemoveServiceAccountFromProject(serviceAccountId, projectId int64) error {
	orgId, err := c.organizationId()
	if err != nil {
		return err
	}

	_, err = c.doJSON("POST", fmt.Sprintf("%s/api/app/organizations/%d/service-accounts/remove-from-project", c.HostURL, orgId), map[string]any{
		"projects": []map[string]any{{"id": projectId, "service_account_ids": []int64{serviceAccountId}}},
	})
	return err
}

// GetProjectServiceAccounts lists the service accounts that are members of a project.
func (c *Client) GetProjectServiceAccounts(projectId int64) ([]ServiceAccountProjectMember, error) {
	// Project endpoints must be called on the project's region, e.g. eu.mixpanel.com.
	project, err := c.GetProject(projectId)
	if err != nil {
		return nil, err
	}
	host := c.HostURL
	if project.Domain == "EU" {
		host = strings.Replace(c.HostURL, "://", "://eu.", 1)
	}

	body, err := c.doJSON("GET", fmt.Sprintf("%s/api/app/projects/%d/service-accounts", host, projectId), nil)
	if err != nil {
		return nil, err
	}

	var response results[[]ServiceAccountProjectMember]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return response.Results, nil
}

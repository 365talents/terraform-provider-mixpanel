package mixpanel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Hardcoded in the Mixpanel frontend code.
const MixpanelUsClusterId = 1
const MixpanelEuClusterId = 5

type Project struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Domain   string `json:"domain"`
	Timezone string `json:"timezone_name"`
	ApiKey   string `json:"api_key"`
	Token    string `json:"token"`
	Secret   string `json:"secret"`
}

type ProjectResponse struct {
	Status  string                 `json:"status"`
	Results ProjectResponseResults `json:"results"`
}

type ProjectResponseResults struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Domain   string `json:"domain"`
	Timezone string `json:"timezone_name"`
	ApiKey   string `json:"api_key"`
	Token    string `json:"token"`
	Secret   string `json:"secret"`
}

func (c *Client) GetProject(id int64) (*Project, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/settings/project/%d/metadata", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var response ProjectResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	project := Project{
		Id:       response.Results.Id,
		Name:     response.Results.Name,
		Timezone: response.Results.Timezone,
		ApiKey:   response.Results.ApiKey,
		Token:    response.Results.Token,
		Secret:   response.Results.Secret,
	}

	if response.Results.Domain == "eu.mixpanel.com" {
		project.Domain = "EU"
	} else {
		project.Domain = "US"
	}

	return &project, nil
}

type createProjectBody struct {
	Name       string `json:"project_name"`
	ClusterId  int64  `json:"cluster_id"`
	TimezoneId int64  `json:"timezone_id"`
}

func (c *Client) CreateProject(project *Project) (*Project, error) {

	var clusterId int64
	if project.Domain == "US" {
		clusterId = MixpanelUsClusterId
	} else {
		clusterId = MixpanelEuClusterId
	}

	timezoneId, err := c.GetTimezoneId(project.Timezone)
	if err != nil {
		return nil, err
	}

	organization, err := c.GetOrganizations()
	if err != nil {
		return nil, err
	}

	// We only support one organization for now
	organizationId := organization[0].Id

	data := createProjectBody{
		Name:       project.Name,
		ClusterId:  clusterId,
		TimezoneId: timezoneId,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/app/organizations/%d/create-project", c.HostURL, organizationId), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var response ProjectResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	project.Id = response.Results.Id
	// log to stdout
	fmt.Println("Project created with ID: ", project.Id)

	return project, nil
}

func (c *Client) UpdateProjectName(id int64, name string) error {
	data := url.Values{}
	data.Set("name", name)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/projects/update/%d", c.HostURL, id), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Referer", c.HostURL)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateProjectTimezone(id int64, timezone string) error {
	data := url.Values{}
	// Don't know if there are cases where timezone_name is different from timezone
	data.Set("timezone", timezone)
	data.Set("timezone_name", timezone)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/projects/update/%d", c.HostURL, id), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Referer", c.HostURL)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

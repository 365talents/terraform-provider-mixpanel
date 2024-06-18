package mixpanel

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type MeResponse struct {
	Status  string    `json:"status"`
	Results MeResults `json:"results"`
}

type MeResults struct {
	Organizations map[string]Organization `json:"organizations"`
}

type Organization struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func (c *Client) GetOrganizations() ([]Organization, error) {
	fmt.Printf("%s", c.HostURL)

	// Not querying the workspace users is a lot faster
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/app/me?include_workspace_users=false", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var response MeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf(err.Error())
		return nil, err
	}

	var orgSlice []Organization
	for _, value := range response.Results.Organizations {
		orgSlice = append(orgSlice, value)
	}

	return orgSlice, nil
}

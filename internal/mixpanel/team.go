package mixpanel

import (
	"encoding/json"
	"fmt"
)

// Teams have no public API, these are the endpoints used by the Mixpanel UI.

type Team struct {
	Id       int64         `json:"id"`
	Name     string        `json:"name"`
	Projects []TeamProject `json:"projects"`
}

type TeamProject struct {
	Id   int64  `json:"id"`
	Role string `json:"role"`
}

// ListTeams fetches the teams once and caches them, so that reading many
// assignments costs one request. Concurrent callers wait for the in-flight
// request. Team writes invalidate the cache.
func (c *Client) ListTeams() ([]Team, error) {
	c.teamsMutex.Lock()
	defer c.teamsMutex.Unlock()
	if c.teams != nil {
		return c.teams, nil
	}

	orgId, err := c.organizationId()
	if err != nil {
		return nil, err
	}

	body, err := c.doJSON("GET", fmt.Sprintf("%s/organizations/%d/teams/", c.HostURL, orgId), nil)
	if err != nil {
		return nil, err
	}

	var response results[[]Team]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	c.teams = response.Results
	return c.teams, nil
}

func (c *Client) invalidateTeams() {
	c.teamsMutex.Lock()
	c.teams = nil
	c.teamsMutex.Unlock()
}

// GetTeam returns nil if the team does not exist.
func (c *Client) GetTeam(id int64) (*Team, error) {
	teams, err := c.ListTeams()
	if err != nil {
		return nil, err
	}
	for _, team := range teams {
		if team.Id == id {
			return &team, nil
		}
	}
	return nil, nil
}

// CreateTeam returns the created team. The endpoint doesn't return the id,
// so the team is found by name afterwards.
func (c *Client) CreateTeam(name string) (*Team, error) {
	orgId, err := c.organizationId()
	if err != nil {
		return nil, err
	}

	body, err := c.doJSON("POST", fmt.Sprintf("%s/organizations/%d/add-teams/", c.HostURL, orgId), map[string]any{"teamNames": []string{name}})
	c.invalidateTeams()
	if err != nil {
		return nil, err
	}

	var response results[[]struct {
		Created bool `json:"created"`
	}]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	if len(response.Results) != 1 || !response.Results[0].Created {
		return nil, fmt.Errorf("team %q was not created, a team with this name may already exist: %s", name, body)
	}

	teams, err := c.ListTeams()
	if err != nil {
		return nil, err
	}
	for _, team := range teams {
		if team.Name == name {
			return &team, nil
		}
	}
	return nil, fmt.Errorf("team %q not found after its creation", name)
}

// AddProjectToTeam adds the project to the team. The resource also calls it to
// change the role, and checks the result.
func (c *Client) AddProjectToTeam(teamId, projectId int64, role string) error {
	orgId, err := c.organizationId()
	if err != nil {
		return err
	}
	defer c.invalidateTeams()

	_, err = c.doJSON("POST", fmt.Sprintf("%s/organizations/%d/add-projects-to-teams/", c.HostURL, orgId), map[string]any{
		"teams":                []map[string]any{{"id": teamId, "projects": []map[string]any{{"id": projectId, "role": role}}}},
		"add_default_projects": false,
	})
	return err
}

func (c *Client) RemoveProjectFromTeam(teamId, projectId int64) error {
	orgId, err := c.organizationId()
	if err != nil {
		return err
	}
	defer c.invalidateTeams()

	_, err = c.doJSON("POST", fmt.Sprintf("%s/organizations/%d/teams/%d/delete-projects/", c.HostURL, orgId, teamId), map[string]any{"projectIds": []int64{projectId}})
	return err
}

func (c *Client) DeleteTeam(id int64) error {
	orgId, err := c.organizationId()
	if err != nil {
		return err
	}
	defer c.invalidateTeams()

	_, err = c.doJSON("POST", fmt.Sprintf("%s/organizations/%d/delete-teams/", c.HostURL, orgId), map[string]any{"teamIds": []int64{id}})
	return err
}

func (c *Client) UpdateTeamName(id int64, name string) error {
	orgId, err := c.organizationId()
	if err != nil {
		return err
	}
	defer c.invalidateTeams()

	_, err = c.doJSON("POST", fmt.Sprintf("%s/organizations/%d/teams/%d/update/", c.HostURL, orgId, id), map[string]any{"name": name})
	return err
}

package mixpanel

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Timezone struct {
	Id			 int64  `json:"id"`
	Name		 string `json:"name"`
}

type TimezoneResponse struct {
	Status string     			`json:"status"`
	Results [][]interface{} `json:"results"`
}

func (c *Client) GetTimezones() ([]Timezone, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/app/timezones", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var response TimezoneResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	timezones := make([]Timezone, 0)
	for _, result := range response.Results {
		timezone := Timezone{
			Id:   int64(result[0].(float64)), // Encoder default to float64 for JSON numbers
			Name: result[1].(string),
		}
		timezones = append(timezones, timezone)
	}

	return timezones, nil
}

func (c *Client) GetTimezoneId(name string) (int64, error) {
	timezones, err := c.GetTimezones()
	if err != nil {
		return 0, err
	}

	for _, timezone := range timezones {
		if timezone.Name == name {
			return timezone.Id, nil
		}
	}

	return 0, fmt.Errorf("Timezone not found: %s", name)
}

func (c *Client) TimezoneIsSupported(name string) (bool, error) {
	timezones, err := c.GetTimezones()
	if err != nil {
		return false, err
	}

	for _, timezone := range timezones {
		if timezone.Name == name {
			return true, nil
		}
	}

	return false, nil
}

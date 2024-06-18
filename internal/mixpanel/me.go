package mixpanel

import (
	"fmt"
	"net/http"
)

func (c *Client) Me() ([]byte, error) {
	fmt.Println("me")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/app/me", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	return body, nil
}

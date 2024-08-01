package mixpanel

import (
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/sync/semaphore"
)

// Default Mixpanel URL.
const HostURL string = "https://mixpanel.com"

type Client struct {
	HostURL    string
	HTTPClient *http.Client
	AuthHeader string
	Semaphore  *semaphore.Weighted
}

func NewClient(serviceAccountUsername, serviceAccountSecret *string, concurrentRequests int64) (*Client, error) {

	retryClient := retryablehttp.NewClient()
	retryClient.Backoff = retryablehttp.DefaultBackoff

	c := Client{
		HTTPClient: retryClient.StandardClient(),
		// Default Hashicups URL
		HostURL: HostURL,
	}

	if serviceAccountUsername == nil || serviceAccountSecret == nil {
		return nil, fmt.Errorf("missing service account credentials")
	}

	c.AuthHeader = "Basic " + *serviceAccountUsername + ":" + *serviceAccountSecret

	c.Semaphore = semaphore.NewWeighted(concurrentRequests)

	return &c, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Add("Authorization", c.AuthHeader)

	err := c.Semaphore.Acquire(req.Context(), 1)
	if err != nil {
		return nil, err
	}
	res, err := c.HTTPClient.Do(req)
	c.Semaphore.Release(1)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}

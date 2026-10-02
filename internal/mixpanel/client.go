package mixpanel

import (
	"fmt"
	"io"
	"net/http"
	"sync"

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

	// Cached for the lifetime of the provider process, see organizationId.
	orgMutex sync.Mutex
	orgId    int64

	// Cached until a team write, see ListTeams.
	teamsMutex sync.Mutex
	teams      []Team

	// Project id to domain ("US" or "EU"), filled by GetProject. A project's domain can't change.
	projectDomains sync.Map
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

// APIError is returned by doRequest for non-2xx responses.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("status: %d, body: %s", e.StatusCode, e.Body)
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
		return nil, &APIError{StatusCode: res.StatusCode, Body: string(body)}
	}

	return body, err
}

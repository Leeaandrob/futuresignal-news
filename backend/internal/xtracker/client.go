// Package xtracker provides integration with the XTracker API for social signal tracking.
package xtracker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// DefaultBaseURL is the XTracker API base URL.
	DefaultBaseURL = "https://xtracker.polymarket.com/api"

	// DefaultTimeout for HTTP requests.
	DefaultTimeout = 30 * time.Second
)

// Client is an XTracker API client.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new XTracker client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL sets a custom base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// GetUsers returns all tracked users.
func (c *Client) GetUsers(ctx context.Context) ([]TrackedUser, error) {
	url := fmt.Sprintf("%s/users", c.baseURL)

	var resp struct {
		Success bool `json:"success"`
		Data    []struct {
			ID         string    `json:"id"`
			Handle     string    `json:"handle"`
			Name       string    `json:"name"`
			PlatformID string    `json:"platformId"`
			AvatarURL  string    `json:"avatarUrl"`
			Bio        string    `json:"bio"`
			Verified   bool      `json:"verified"`
			CreatedAt  time.Time `json:"createdAt"`
			UpdatedAt  time.Time `json:"updatedAt"`
			Count      struct {
				Posts int `json:"posts"`
			} `json:"_count"`
		} `json:"data"`
	}

	if err := c.doRequest(ctx, url, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	users := make([]TrackedUser, len(resp.Data))
	for i, u := range resp.Data {
		users[i] = TrackedUser{
			ID:         u.ID,
			Handle:     u.Handle,
			Name:       u.Name,
			PlatformID: u.PlatformID,
			AvatarURL:  u.AvatarURL,
			Bio:        u.Bio,
			Verified:   u.Verified,
			CreatedAt:  u.CreatedAt,
			UpdatedAt:  u.UpdatedAt,
			PostCount:  u.Count.Posts,
		}
	}

	return users, nil
}

// GetUser returns a specific user by handle.
func (c *Client) GetUser(ctx context.Context, handle string) (*TrackedUser, error) {
	url := fmt.Sprintf("%s/users/%s", c.baseURL, handle)

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			ID         string    `json:"id"`
			Handle     string    `json:"handle"`
			Name       string    `json:"name"`
			PlatformID string    `json:"platformId"`
			AvatarURL  string    `json:"avatarUrl"`
			Bio        string    `json:"bio"`
			Verified   bool      `json:"verified"`
			CreatedAt  time.Time `json:"createdAt"`
			UpdatedAt  time.Time `json:"updatedAt"`
		} `json:"data"`
	}

	if err := c.doRequest(ctx, url, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return &TrackedUser{
		ID:         resp.Data.ID,
		Handle:     resp.Data.Handle,
		Name:       resp.Data.Name,
		PlatformID: resp.Data.PlatformID,
		AvatarURL:  resp.Data.AvatarURL,
		Bio:        resp.Data.Bio,
		Verified:   resp.Data.Verified,
		CreatedAt:  resp.Data.CreatedAt,
		UpdatedAt:  resp.Data.UpdatedAt,
	}, nil
}

// GetPosts returns posts for a user, optionally limited.
func (c *Client) GetPosts(ctx context.Context, handle string, limit int) ([]Post, error) {
	url := fmt.Sprintf("%s/users/%s/posts", c.baseURL, handle)
	if limit > 0 {
		url = fmt.Sprintf("%s?limit=%d", url, limit)
	}

	var resp PostsResponse
	if err := c.doRequest(ctx, url, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return resp.Data, nil
}

// GetRecentPosts returns posts from a user since a given time.
func (c *Client) GetRecentPosts(ctx context.Context, handle string, since time.Time, limit int) ([]Post, error) {
	posts, err := c.GetPosts(ctx, handle, limit)
	if err != nil {
		return nil, err
	}

	// Filter by time
	var recent []Post
	for _, p := range posts {
		if p.CreatedAt.After(since) {
			recent = append(recent, p)
		}
	}

	return recent, nil
}

// GetActiveTrackings returns all active tracking periods.
func (c *Client) GetActiveTrackings(ctx context.Context) ([]Tracking, error) {
	url := fmt.Sprintf("%s/trackings?activeOnly=true", c.baseURL)

	var resp TrackingsResponse
	if err := c.doRequest(ctx, url, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return resp.Data, nil
}

// GetMetrics returns metrics for a user within a date range.
func (c *Client) GetMetrics(ctx context.Context, userID string, startDate, endDate time.Time) ([]DailyMetric, error) {
	url := fmt.Sprintf("%s/metrics/%s?type=daily&startDate=%s&endDate=%s",
		c.baseURL, userID,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))

	var resp MetricsResponse
	if err := c.doRequest(ctx, url, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return resp.Data, nil
}

// doRequest performs an HTTP GET request and decodes the JSON response.
func (c *Client) doRequest(ctx context.Context, url string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "FutureSignals/1.0")

	log.Debug().Str("url", url).Msg("XTracker API request")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	return nil
}

// HealthCheck verifies the API is accessible.
func (c *Client) HealthCheck(ctx context.Context) error {
	users, err := c.GetUsers(ctx)
	if err != nil {
		return err
	}
	log.Info().Int("tracked_users", len(users)).Msg("XTracker API health check passed")
	return nil
}

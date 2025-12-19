// Package enrichment provides context enrichment for signal narratives.
package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

const (
	TavilyAPIURL = "https://api.tavily.com"
)

// TavilyClient provides search functionality via Tavily API.
type TavilyClient struct {
	client *resty.Client
	apiKey string
}

// TavilySearchRequest represents a search request.
type TavilySearchRequest struct {
	Query             string   `json:"query"`
	SearchDepth       string   `json:"search_depth,omitempty"`       // "basic" or "advanced"
	Topic             string   `json:"topic,omitempty"`              // "general" or "news"
	MaxResults        int      `json:"max_results,omitempty"`
	IncludeAnswer     bool     `json:"include_answer,omitempty"`
	IncludeRawContent bool     `json:"include_raw_content,omitempty"`
	IncludeDomains    []string `json:"include_domains,omitempty"`
	ExcludeDomains    []string `json:"exclude_domains,omitempty"`
}

// TavilySearchResponse represents a search response.
type TavilySearchResponse struct {
	Query   string         `json:"query"`
	Answer  string         `json:"answer,omitempty"`
	Results []TavilyResult `json:"results"`
}

// TavilyResult represents a single search result.
type TavilyResult struct {
	Title      string  `json:"title"`
	URL        string  `json:"url"`
	Content    string  `json:"content"`
	RawContent string  `json:"raw_content,omitempty"`
	Score      float64 `json:"score"`
	Published  string  `json:"published_date,omitempty"`
}

// NewTavilyClient creates a new Tavily client.
func NewTavilyClient(apiKey string) *TavilyClient {
	return &TavilyClient{
		client: resty.New().
			SetBaseURL(TavilyAPIURL).
			SetTimeout(30 * time.Second).
			SetRetryCount(2),
		apiKey: apiKey,
	}
}

// Search performs a search query.
func (c *TavilyClient) Search(ctx context.Context, query string, maxResults int) (*TavilySearchResponse, error) {
	return c.SearchAdvanced(ctx, TavilySearchRequest{
		Query:         query,
		SearchDepth:   "basic",
		Topic:         "news",
		MaxResults:    maxResults,
		IncludeAnswer: true,
	})
}

// SearchNews performs a news-focused search.
func (c *TavilyClient) SearchNews(ctx context.Context, query string, maxResults int) (*TavilySearchResponse, error) {
	return c.SearchAdvanced(ctx, TavilySearchRequest{
		Query:         query,
		SearchDepth:   "advanced",
		Topic:         "news",
		MaxResults:    maxResults,
		IncludeAnswer: true,
		IncludeDomains: []string{
			"reuters.com",
			"bloomberg.com",
			"cnbc.com",
			"wsj.com",
			"ft.com",
			"bbc.com",
			"cnn.com",
			"apnews.com",
		},
	})
}

// SearchAdvanced performs a search with custom parameters.
func (c *TavilyClient) SearchAdvanced(ctx context.Context, req TavilySearchRequest) (*TavilySearchResponse, error) {
	body := map[string]interface{}{
		"api_key":         c.apiKey,
		"query":           req.Query,
		"search_depth":    req.SearchDepth,
		"topic":           req.Topic,
		"max_results":     req.MaxResults,
		"include_answer":  req.IncludeAnswer,
	}

	if len(req.IncludeDomains) > 0 {
		body["include_domains"] = req.IncludeDomains
	}
	if len(req.ExcludeDomains) > 0 {
		body["exclude_domains"] = req.ExcludeDomains
	}
	if req.IncludeRawContent {
		body["include_raw_content"] = true
	}

	log.Debug().
		Str("query", req.Query).
		Int("max_results", req.MaxResults).
		Msg("Tavily search")

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("/search")

	if err != nil {
		return nil, fmt.Errorf("tavily search failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("tavily API returned %d: %s", resp.StatusCode(), resp.String())
	}

	var result TavilySearchResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse tavily response: %w", err)
	}

	log.Debug().
		Int("results", len(result.Results)).
		Bool("has_answer", result.Answer != "").
		Msg("Tavily search complete")

	return &result, nil
}

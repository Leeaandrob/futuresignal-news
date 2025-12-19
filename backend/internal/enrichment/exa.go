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
	ExaAPIURL = "https://api.exa.ai"
)

// ExaClient provides semantic search functionality via Exa API.
type ExaClient struct {
	client *resty.Client
	apiKey string
}

// ExaSearchRequest represents a search request.
type ExaSearchRequest struct {
	Query            string    `json:"query"`
	Type             string    `json:"type,omitempty"`               // "keyword", "neural", "auto"
	UseAutoprompt    bool      `json:"useAutoprompt,omitempty"`
	NumResults       int       `json:"numResults,omitempty"`
	StartCrawlDate   string    `json:"startCrawlDate,omitempty"`     // ISO date
	EndCrawlDate     string    `json:"endCrawlDate,omitempty"`
	StartPublishDate string    `json:"startPublishedDate,omitempty"` // ISO date
	EndPublishDate   string    `json:"endPublishedDate,omitempty"`
	IncludeDomains   []string  `json:"includeDomains,omitempty"`
	ExcludeDomains   []string  `json:"excludeDomains,omitempty"`
	Category         string    `json:"category,omitempty"`           // "news", "company", "research_paper", etc.
	Contents         *ExaContents `json:"contents,omitempty"`
}

// ExaContents specifies what content to return.
type ExaContents struct {
	Text      *ExaTextOptions      `json:"text,omitempty"`
	Highlights *ExaHighlightOptions `json:"highlights,omitempty"`
	Summary   *ExaSummaryOptions   `json:"summary,omitempty"`
}

// ExaTextOptions specifies text extraction options.
type ExaTextOptions struct {
	MaxCharacters     int  `json:"maxCharacters,omitempty"`
	IncludeHTMLTags   bool `json:"includeHtmlTags,omitempty"`
}

// ExaHighlightOptions specifies highlight extraction options.
type ExaHighlightOptions struct {
	NumSentences      int    `json:"numSentences,omitempty"`
	HighlightsPerURL  int    `json:"highlightsPerUrl,omitempty"`
	Query             string `json:"query,omitempty"`
}

// ExaSummaryOptions specifies summary options.
type ExaSummaryOptions struct {
	Query string `json:"query,omitempty"`
}

// ExaSearchResponse represents a search response.
type ExaSearchResponse struct {
	Results           []ExaResult `json:"results"`
	AutopromptString  string      `json:"autopromptString,omitempty"`
}

// ExaResult represents a single search result.
type ExaResult struct {
	ID            string   `json:"id"`
	URL           string   `json:"url"`
	Title         string   `json:"title"`
	Score         float64  `json:"score,omitempty"`
	PublishedDate string   `json:"publishedDate,omitempty"`
	Author        string   `json:"author,omitempty"`
	Text          string   `json:"text,omitempty"`
	Highlights    []string `json:"highlights,omitempty"`
	Summary       string   `json:"summary,omitempty"`
}

// NewExaClient creates a new Exa client.
func NewExaClient(apiKey string) *ExaClient {
	return &ExaClient{
		client: resty.New().
			SetBaseURL(ExaAPIURL).
			SetTimeout(30 * time.Second).
			SetRetryCount(2),
		apiKey: apiKey,
	}
}

// Search performs a semantic search query.
func (c *ExaClient) Search(ctx context.Context, query string, numResults int) (*ExaSearchResponse, error) {
	return c.SearchAdvanced(ctx, ExaSearchRequest{
		Query:         query,
		Type:          "auto",
		UseAutoprompt: true,
		NumResults:    numResults,
		Contents: &ExaContents{
			Text: &ExaTextOptions{
				MaxCharacters: 1000,
			},
			Highlights: &ExaHighlightOptions{
				NumSentences:     3,
				HighlightsPerURL: 3,
			},
		},
	})
}

// SearchNews performs a news-focused semantic search.
func (c *ExaClient) SearchNews(ctx context.Context, query string, numResults int, daysBack int) (*ExaSearchResponse, error) {
	startDate := time.Now().AddDate(0, 0, -daysBack).Format("2006-01-02")

	return c.SearchAdvanced(ctx, ExaSearchRequest{
		Query:            query,
		Type:             "neural",
		UseAutoprompt:    true,
		NumResults:       numResults,
		Category:         "news",
		StartPublishDate: startDate,
		Contents: &ExaContents{
			Text: &ExaTextOptions{
				MaxCharacters: 1500,
			},
			Highlights: &ExaHighlightOptions{
				NumSentences:     3,
				HighlightsPerURL: 3,
				Query:            query,
			},
			Summary: &ExaSummaryOptions{
				Query: query,
			},
		},
	})
}

// SearchAdvanced performs a search with custom parameters.
func (c *ExaClient) SearchAdvanced(ctx context.Context, req ExaSearchRequest) (*ExaSearchResponse, error) {
	log.Debug().
		Str("query", req.Query).
		Int("num_results", req.NumResults).
		Str("type", req.Type).
		Msg("Exa search")

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("x-api-key", c.apiKey).
		SetBody(req).
		Post("/search")

	if err != nil {
		return nil, fmt.Errorf("exa search failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("exa API returned %d: %s", resp.StatusCode(), resp.String())
	}

	var result ExaSearchResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse exa response: %w", err)
	}

	log.Debug().
		Int("results", len(result.Results)).
		Str("autoprompt", result.AutopromptString).
		Msg("Exa search complete")

	return &result, nil
}

// FindSimilar finds content similar to a given URL.
func (c *ExaClient) FindSimilar(ctx context.Context, url string, numResults int) (*ExaSearchResponse, error) {
	body := map[string]interface{}{
		"url":        url,
		"numResults": numResults,
		"contents": map[string]interface{}{
			"text": map[string]interface{}{
				"maxCharacters": 1000,
			},
			"highlights": map[string]interface{}{
				"numSentences":     3,
				"highlightsPerUrl": 3,
			},
		},
	}

	log.Debug().
		Str("url", url).
		Int("num_results", numResults).
		Msg("Exa find similar")

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("x-api-key", c.apiKey).
		SetBody(body).
		Post("/findSimilar")

	if err != nil {
		return nil, fmt.Errorf("exa find similar failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("exa API returned %d: %s", resp.StatusCode(), resp.String())
	}

	var result ExaSearchResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse exa response: %w", err)
	}

	return &result, nil
}

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
	FirecrawlAPIURL = "https://api.firecrawl.dev/v1"
)

// FirecrawlClient provides web scraping functionality via Firecrawl API.
type FirecrawlClient struct {
	client *resty.Client
	apiKey string
}

// FirecrawlScrapeRequest represents a scrape request.
type FirecrawlScrapeRequest struct {
	URL     string   `json:"url"`
	Formats []string `json:"formats,omitempty"` // "markdown", "html", "rawHtml", "links", "screenshot"
}

// FirecrawlScrapeResponse represents a scrape response.
type FirecrawlScrapeResponse struct {
	Success bool                 `json:"success"`
	Data    *FirecrawlScrapeData `json:"data,omitempty"`
	Error   string               `json:"error,omitempty"`
}

// FirecrawlScrapeData represents scraped page data.
type FirecrawlScrapeData struct {
	Markdown string                 `json:"markdown,omitempty"`
	HTML     string                 `json:"html,omitempty"`
	RawHTML  string                 `json:"rawHtml,omitempty"`
	Links    []string               `json:"links,omitempty"`
	Metadata FirecrawlPageMetadata  `json:"metadata,omitempty"`
}

// FirecrawlPageMetadata represents page metadata.
type FirecrawlPageMetadata struct {
	Title         string `json:"title,omitempty"`
	Description   string `json:"description,omitempty"`
	Language      string `json:"language,omitempty"`
	OGTitle       string `json:"ogTitle,omitempty"`
	OGDescription string `json:"ogDescription,omitempty"`
	OGImage       string `json:"ogImage,omitempty"`
	OGUrl         string `json:"ogUrl,omitempty"`
	SourceURL     string `json:"sourceURL,omitempty"`
}

// NewFirecrawlClient creates a new Firecrawl client.
func NewFirecrawlClient(apiKey string) *FirecrawlClient {
	return &FirecrawlClient{
		client: resty.New().
			SetBaseURL(FirecrawlAPIURL).
			SetTimeout(60 * time.Second).
			SetRetryCount(2),
		apiKey: apiKey,
	}
}

// Scrape extracts content from a URL.
func (c *FirecrawlClient) Scrape(ctx context.Context, url string) (*FirecrawlScrapeData, error) {
	return c.ScrapeWithFormats(ctx, url, []string{"markdown"})
}

// ScrapeWithFormats extracts content from a URL with specific formats.
func (c *FirecrawlClient) ScrapeWithFormats(ctx context.Context, url string, formats []string) (*FirecrawlScrapeData, error) {
	body := map[string]interface{}{
		"url":     url,
		"formats": formats,
	}

	log.Debug().
		Str("url", url).
		Strs("formats", formats).
		Msg("Firecrawl scrape")

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", c.apiKey)).
		SetBody(body).
		Post("/scrape")

	if err != nil {
		return nil, fmt.Errorf("firecrawl scrape failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("firecrawl API returned %d: %s", resp.StatusCode(), resp.String())
	}

	var result FirecrawlScrapeResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse firecrawl response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("firecrawl scrape failed: %s", result.Error)
	}

	log.Debug().
		Str("title", result.Data.Metadata.Title).
		Int("markdown_len", len(result.Data.Markdown)).
		Msg("Firecrawl scrape complete")

	return result.Data, nil
}

// ScrapeMultiple extracts content from multiple URLs concurrently.
func (c *FirecrawlClient) ScrapeMultiple(ctx context.Context, urls []string, maxConcurrent int) ([]*FirecrawlScrapeData, error) {
	if maxConcurrent <= 0 {
		maxConcurrent = 3
	}

	results := make([]*FirecrawlScrapeData, len(urls))
	errors := make([]error, len(urls))

	// Simple sequential scraping with limit
	// Could be enhanced with goroutines for parallel execution
	for i, url := range urls {
		if i >= maxConcurrent*2 { // Limit total URLs
			break
		}

		data, err := c.Scrape(ctx, url)
		if err != nil {
			log.Warn().Err(err).Str("url", url).Msg("Failed to scrape URL")
			errors[i] = err
			continue
		}
		results[i] = data
	}

	// Filter out nil results
	var validResults []*FirecrawlScrapeData
	for _, r := range results {
		if r != nil {
			validResults = append(validResults, r)
		}
	}

	return validResults, nil
}

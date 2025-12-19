// Package enrichment provides context enrichment for signal narratives.
package enrichment

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// EnrichmentConfig holds configuration for the enricher.
type EnrichmentConfig struct {
	TavilyAPIKey    string
	ExaAPIKey       string
	FirecrawlAPIKey string
	MaxNewsResults  int
	MaxDeepScrapes  int
	EnableTavily    bool
	EnableExa       bool
	EnableFirecrawl bool
}

// Enricher orchestrates context enrichment from multiple sources.
type Enricher struct {
	tavily    *TavilyClient
	exa       *ExaClient
	firecrawl *FirecrawlClient
	config    EnrichmentConfig
}

// EnrichedContext represents the combined context from all sources.
type EnrichedContext struct {
	// News articles from Tavily
	NewsArticles []NewsArticle `json:"news_articles"`

	// Semantic search results from Exa
	SemanticResults []SemanticResult `json:"semantic_results"`

	// Deep scraped content from Firecrawl
	DeepContent []DeepContent `json:"deep_content"`

	// Combined summary for LLM consumption
	Summary string `json:"summary"`

	// Metadata
	EnrichedAt time.Time `json:"enriched_at"`
	Sources    []string  `json:"sources"`
}

// NewsArticle represents a news article from Tavily.
type NewsArticle struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Content     string    `json:"content"`
	Published   string    `json:"published,omitempty"`
	Source      string    `json:"source"`
	Relevance   float64   `json:"relevance"`
}

// SemanticResult represents a semantic search result from Exa.
type SemanticResult struct {
	Title      string   `json:"title"`
	URL        string   `json:"url"`
	Text       string   `json:"text,omitempty"`
	Highlights []string `json:"highlights,omitempty"`
	Summary    string   `json:"summary,omitempty"`
	Published  string   `json:"published,omitempty"`
	Score      float64  `json:"score"`
}

// DeepContent represents deeply scraped content from Firecrawl.
type DeepContent struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Markdown    string `json:"markdown"`
	Description string `json:"description,omitempty"`
}

// NewEnricher creates a new Enricher with the given configuration.
func NewEnricher(config EnrichmentConfig) *Enricher {
	e := &Enricher{
		config: config,
	}

	if config.EnableTavily && config.TavilyAPIKey != "" {
		e.tavily = NewTavilyClient(config.TavilyAPIKey)
		log.Info().Msg("Tavily enrichment enabled")
	}

	if config.EnableExa && config.ExaAPIKey != "" {
		e.exa = NewExaClient(config.ExaAPIKey)
		log.Info().Msg("Exa enrichment enabled")
	}

	if config.EnableFirecrawl && config.FirecrawlAPIKey != "" {
		e.firecrawl = NewFirecrawlClient(config.FirecrawlAPIKey)
		log.Info().Msg("Firecrawl enrichment enabled")
	}

	if config.MaxNewsResults <= 0 {
		e.config.MaxNewsResults = 5
	}
	if config.MaxDeepScrapes <= 0 {
		e.config.MaxDeepScrapes = 2
	}

	return e
}

// Enrich gathers context for a market signal from multiple sources.
func (e *Enricher) Enrich(ctx context.Context, marketQuestion string, category string) (*EnrichedContext, error) {
	log.Info().
		Str("market", marketQuestion).
		Str("category", category).
		Msg("Starting enrichment")

	result := &EnrichedContext{
		EnrichedAt: time.Now(),
		Sources:    []string{},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	errs := make([]error, 0)

	// Run all enrichment sources concurrently
	if e.tavily != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			articles, err := e.enrichFromTavily(ctx, marketQuestion)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				log.Warn().Err(err).Msg("Tavily enrichment failed")
				errs = append(errs, err)
			} else {
				result.NewsArticles = articles
				result.Sources = append(result.Sources, "tavily")
			}
		}()
	}

	if e.exa != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			semantic, err := e.enrichFromExa(ctx, marketQuestion, category)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				log.Warn().Err(err).Msg("Exa enrichment failed")
				errs = append(errs, err)
			} else {
				result.SemanticResults = semantic
				result.Sources = append(result.Sources, "exa")
			}
		}()
	}

	wg.Wait()

	// Deep scrape top URLs if Firecrawl is enabled
	if e.firecrawl != nil && len(result.NewsArticles) > 0 {
		deepContent, err := e.enrichWithFirecrawl(ctx, result)
		if err != nil {
			log.Warn().Err(err).Msg("Firecrawl enrichment failed")
		} else {
			result.DeepContent = deepContent
			result.Sources = append(result.Sources, "firecrawl")
		}
	}

	// Generate combined summary
	result.Summary = e.generateSummary(result, marketQuestion)

	log.Info().
		Int("news_articles", len(result.NewsArticles)).
		Int("semantic_results", len(result.SemanticResults)).
		Int("deep_content", len(result.DeepContent)).
		Strs("sources", result.Sources).
		Msg("Enrichment complete")

	return result, nil
}

// enrichFromTavily fetches news articles from Tavily.
func (e *Enricher) enrichFromTavily(ctx context.Context, query string) ([]NewsArticle, error) {
	resp, err := e.tavily.SearchNews(ctx, query, e.config.MaxNewsResults)
	if err != nil {
		return nil, err
	}

	articles := make([]NewsArticle, 0, len(resp.Results))
	for _, r := range resp.Results {
		// Extract domain from URL as source
		source := extractDomain(r.URL)

		articles = append(articles, NewsArticle{
			Title:     r.Title,
			URL:       r.URL,
			Content:   r.Content,
			Published: r.Published,
			Source:    source,
			Relevance: r.Score,
		})
	}

	return articles, nil
}

// enrichFromExa fetches semantic search results from Exa.
func (e *Enricher) enrichFromExa(ctx context.Context, query string, category string) ([]SemanticResult, error) {
	// Search for recent news related to the query
	resp, err := e.exa.SearchNews(ctx, query, e.config.MaxNewsResults, 7) // Last 7 days
	if err != nil {
		return nil, err
	}

	results := make([]SemanticResult, 0, len(resp.Results))
	for _, r := range resp.Results {
		results = append(results, SemanticResult{
			Title:      r.Title,
			URL:        r.URL,
			Text:       r.Text,
			Highlights: r.Highlights,
			Summary:    r.Summary,
			Published:  r.PublishedDate,
			Score:      r.Score,
		})
	}

	return results, nil
}

// enrichWithFirecrawl deep scrapes the top URLs for detailed content.
func (e *Enricher) enrichWithFirecrawl(ctx context.Context, enriched *EnrichedContext) ([]DeepContent, error) {
	// Collect top URLs from news articles
	urls := make([]string, 0)
	for _, article := range enriched.NewsArticles {
		if len(urls) >= e.config.MaxDeepScrapes {
			break
		}
		urls = append(urls, article.URL)
	}

	if len(urls) == 0 {
		return nil, nil
	}

	scraped, err := e.firecrawl.ScrapeMultiple(ctx, urls, e.config.MaxDeepScrapes)
	if err != nil {
		return nil, err
	}

	content := make([]DeepContent, 0, len(scraped))
	for _, s := range scraped {
		if s == nil {
			continue
		}
		content = append(content, DeepContent{
			Title:       s.Metadata.Title,
			URL:         s.Metadata.SourceURL,
			Markdown:    truncateString(s.Markdown, 3000), // Limit for LLM context
			Description: s.Metadata.Description,
		})
	}

	return content, nil
}

// generateSummary creates a combined summary for LLM consumption.
func (e *Enricher) generateSummary(enriched *EnrichedContext, query string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== CONTEXT FOR: %s ===\n\n", query))

	if len(enriched.NewsArticles) > 0 {
		sb.WriteString("## Recent News:\n")
		for i, article := range enriched.NewsArticles {
			sb.WriteString(fmt.Sprintf("%d. **%s** (%s)\n", i+1, article.Title, article.Source))
			if article.Content != "" {
				sb.WriteString(fmt.Sprintf("   %s\n", truncateString(article.Content, 300)))
			}
			sb.WriteString("\n")
		}
	}

	if len(enriched.SemanticResults) > 0 {
		sb.WriteString("\n## Related Analysis:\n")
		for i, result := range enriched.SemanticResults {
			sb.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, result.Title))
			if result.Summary != "" {
				sb.WriteString(fmt.Sprintf("   Summary: %s\n", result.Summary))
			}
			if len(result.Highlights) > 0 {
				sb.WriteString("   Key Points:\n")
				for _, h := range result.Highlights[:min(3, len(result.Highlights))] {
					sb.WriteString(fmt.Sprintf("   - %s\n", h))
				}
			}
			sb.WriteString("\n")
		}
	}

	if len(enriched.DeepContent) > 0 {
		sb.WriteString("\n## Detailed Sources:\n")
		for i, content := range enriched.DeepContent {
			sb.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, content.Title))
			if content.Description != "" {
				sb.WriteString(fmt.Sprintf("   %s\n", content.Description))
			}
			// Include truncated markdown for deep context
			if content.Markdown != "" {
				sb.WriteString(fmt.Sprintf("\n   --- Excerpt ---\n   %s\n   ---\n\n", truncateString(content.Markdown, 1000)))
			}
		}
	}

	return sb.String()
}

// Helper functions

func extractDomain(url string) string {
	// Simple domain extraction
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "www.")
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return url
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

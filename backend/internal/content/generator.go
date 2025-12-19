// Package content provides article generation for FutureSignals.
package content

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/leeaandrob/futuresignals/internal/enrichment"
	"github.com/leeaandrob/futuresignals/internal/models"
	"github.com/leeaandrob/futuresignals/internal/qwen"
	"github.com/leeaandrob/futuresignals/internal/storage"
	"github.com/leeaandrob/futuresignals/internal/sync"
	"github.com/rs/zerolog/log"
)

// Generator creates articles from market data.
type Generator struct {
	store    *storage.Store
	syncer   *sync.Syncer
	llm      *qwen.Client
	enricher *enrichment.Enricher
}

// NewGenerator creates a new content generator.
func NewGenerator(store *storage.Store, syncer *sync.Syncer, llm *qwen.Client, enricher *enrichment.Enricher) *Generator {
	return &Generator{
		store:    store,
		syncer:   syncer,
		llm:      llm,
		enricher: enricher,
	}
}

// GenerateBreaking generates a breaking news article from a market event.
func (g *Generator) GenerateBreaking(ctx context.Context, event sync.Event) (*models.Article, error) {
	log.Info().
		Str("market", event.Market.Question).
		Str("type", string(event.Type)).
		Msg("Generating breaking article")

	// Enrich context
	enrichedCtx := ""
	var sources []string
	if g.enricher != nil {
		ctx, err := g.enricher.Enrich(ctx, event.Market.Question, event.Market.Category)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to enrich context")
		} else if ctx != nil {
			enrichedCtx = ctx.Summary
			sources = ctx.Sources
		}
	}

	// Generate narrative with LLM
	narrative, err := g.generateNarrative(ctx, event.Market, enrichedCtx, "breaking")
	if err != nil {
		return nil, fmt.Errorf("failed to generate narrative: %w", err)
	}

	// Create article
	article := &models.Article{
		Slug:        g.generateSlug(narrative.Headline),
		Type:        models.ArticleTypeBreaking,
		Category:    event.Market.Category,
		Headline:    narrative.Headline,
		Subheadline: narrative.Subheadline,
		Summary:     narrative.Subheadline,
		Body: models.ArticleBody{
			WhatHappened: narrative.WhatChanged,
			WhyItMatters: narrative.WhyItMatters,
			Context:      []string{narrative.MarketContext},
			WhatToWatch:  narrative.WhatToWatch,
		},
		Markets: []models.MarketRef{{
			MarketID:     event.Market.MarketID,
			Question:     event.Market.Question,
			Slug:         event.Market.Slug,
			Probability:  event.Market.Probability,
			PreviousProb: event.Market.PreviousProb,
			Change24h:    event.Market.Change24h,
			Volume24h:    event.Market.Volume24h,
			TotalVolume:  event.Market.TotalVolume,
		}},
		PrimaryMarket: &models.MarketRef{
			MarketID:    event.Market.MarketID,
			Question:    event.Market.Question,
			Probability: event.Market.Probability,
			Change24h:   event.Market.Change24h,
			Volume24h:   event.Market.Volume24h,
		},
		Tags:              narrative.Tags,
		Significance:      models.Significance(narrative.Significance),
		Sentiment:         narrative.Sentiment,
		MetaTitle:         narrative.Headline,
		MetaDescription:   narrative.Subheadline,
		Published:         true,
		EnrichmentSources: sources,
	}

	// Save to database
	if err := g.store.SaveArticle(ctx, article); err != nil {
		return nil, fmt.Errorf("failed to save article: %w", err)
	}

	log.Info().
		Str("slug", article.Slug).
		Str("headline", article.Headline).
		Msg("Breaking article generated")

	return article, nil
}

// GenerateBriefing generates a scheduled briefing article.
func (g *Generator) GenerateBriefing(ctx context.Context, briefingType models.BriefingType) (*models.Article, error) {
	config := models.DefaultBriefingConfigs[briefingType]

	log.Info().
		Str("type", string(briefingType)).
		Str("title", config.Title).
		Msg("Generating briefing")

	// Collect top markets per category
	var allMarkets []models.MarketRef
	for _, category := range config.Categories {
		markets, err := g.store.GetMarketsByCategory(ctx, category, config.MarketsPerCat)
		if err != nil {
			log.Warn().Err(err).Str("category", category).Msg("Failed to get markets")
			continue
		}

		for _, m := range markets {
			allMarkets = append(allMarkets, models.MarketRef{
				MarketID:    m.MarketID,
				Question:    m.Question,
				Slug:        m.Slug,
				Probability: m.Probability,
				Change24h:   m.Change24h,
				Volume24h:   m.Volume24h,
				TotalVolume: m.TotalVolume,
			})
		}
	}

	if len(allMarkets) == 0 {
		return nil, fmt.Errorf("no markets found for briefing")
	}

	// Generate briefing content with LLM
	briefingContent, err := g.generateBriefingContent(ctx, briefingType, allMarkets)
	if err != nil {
		return nil, fmt.Errorf("failed to generate briefing content: %w", err)
	}

	// Create article
	now := time.Now()
	dateStr := now.Format("January 2, 2006")
	slug := fmt.Sprintf("%s-briefing-%s", strings.ToLower(string(briefingType)), now.Format("2006-01-02"))

	article := &models.Article{
		Slug:        slug,
		Type:        models.ArticleTypeBriefing,
		Category:    "briefing",
		Headline:    fmt.Sprintf("%s: %s", config.Title, dateStr),
		Subheadline: briefingContent.Summary,
		Summary:     briefingContent.Summary,
		Body: models.ArticleBody{
			WhatHappened: briefingContent.Overview,
			WhyItMatters: briefingContent.KeyInsights,
			Context:      briefingContent.Highlights,
			WhatToWatch:  briefingContent.WhatToWatch,
		},
		Markets:         allMarkets,
		Tags:            []string{"briefing", string(briefingType), "daily", "markets"},
		Significance:    models.SignificanceMedium,
		Sentiment:       "neutral",
		MetaTitle:       fmt.Sprintf("%s - %s | FutureSignals", config.Title, dateStr),
		MetaDescription: briefingContent.Summary,
		Published:       true,
	}

	if err := g.store.SaveArticle(ctx, article); err != nil {
		return nil, fmt.Errorf("failed to save article: %w", err)
	}

	log.Info().
		Str("slug", article.Slug).
		Int("markets", len(allMarkets)).
		Msg("Briefing generated")

	return article, nil
}

// GenerateTrending generates an article about trending markets.
func (g *Generator) GenerateTrending(ctx context.Context, limit int) (*models.Article, error) {
	log.Info().Int("limit", limit).Msg("Generating trending article")

	// Get trending markets
	markets, err := g.store.GetTrendingMarkets(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending markets: %w", err)
	}

	if len(markets) == 0 {
		return nil, fmt.Errorf("no trending markets found")
	}

	// Convert to refs
	var marketRefs []models.MarketRef
	for _, m := range markets {
		marketRefs = append(marketRefs, models.MarketRef{
			MarketID:    m.MarketID,
			Question:    m.Question,
			Slug:        m.Slug,
			Probability: m.Probability,
			Change24h:   m.Change24h,
			Volume24h:   m.Volume24h,
			TotalVolume: m.TotalVolume,
		})
	}

	// Generate content
	trendingContent, err := g.generateTrendingContent(ctx, marketRefs)
	if err != nil {
		return nil, fmt.Errorf("failed to generate trending content: %w", err)
	}

	now := time.Now()
	slug := fmt.Sprintf("trending-markets-%s", now.Format("2006-01-02-1504"))

	article := &models.Article{
		Slug:        slug,
		Type:        models.ArticleTypeTrending,
		Category:    "trending",
		Headline:    trendingContent.Headline,
		Subheadline: trendingContent.Summary,
		Summary:     trendingContent.Summary,
		Body: models.ArticleBody{
			WhatHappened: trendingContent.Overview,
			WhyItMatters: trendingContent.Analysis,
			Context:      trendingContent.Highlights,
			WhatToWatch:  trendingContent.WhatToWatch,
		},
		Markets:         marketRefs,
		Tags:            append([]string{"trending", "hot", "markets"}, trendingContent.Tags...),
		Significance:    models.SignificanceMedium,
		Sentiment:       "neutral",
		MetaTitle:       trendingContent.Headline + " | FutureSignals",
		MetaDescription: trendingContent.Summary,
		Published:       true,
	}

	if err := g.store.SaveArticle(ctx, article); err != nil {
		return nil, fmt.Errorf("failed to save article: %w", err)
	}

	log.Info().
		Str("slug", article.Slug).
		Int("markets", len(marketRefs)).
		Msg("Trending article generated")

	return article, nil
}

// GenerateNewMarket generates an article about a new market.
func (g *Generator) GenerateNewMarket(ctx context.Context, market *models.Market) (*models.Article, error) {
	log.Info().
		Str("market", market.Question).
		Msg("Generating new market article")

	// Enrich context
	enrichedCtx := ""
	var sources []string
	if g.enricher != nil {
		ctx, err := g.enricher.Enrich(ctx, market.Question, market.Category)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to enrich context")
		} else if ctx != nil {
			enrichedCtx = ctx.Summary
			sources = ctx.Sources
		}
	}

	// Generate content
	content, err := g.generateNewMarketContent(ctx, market, enrichedCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	slug := fmt.Sprintf("new-market-%s-%s", market.Slug, time.Now().Format("20060102"))

	article := &models.Article{
		Slug:        slug,
		Type:        models.ArticleTypeNewMarket,
		Category:    market.Category,
		Headline:    content.Headline,
		Subheadline: content.Summary,
		Summary:     content.Summary,
		Body: models.ArticleBody{
			WhatHappened: content.Overview,
			WhyItMatters: content.WhyItMatters,
			Context:      content.Context,
			WhatToWatch:  content.WhatToWatch,
		},
		Markets: []models.MarketRef{{
			MarketID:    market.MarketID,
			Question:    market.Question,
			Slug:        market.Slug,
			Probability: market.Probability,
			Volume24h:   market.Volume24h,
			TotalVolume: market.TotalVolume,
		}},
		PrimaryMarket: &models.MarketRef{
			MarketID:    market.MarketID,
			Question:    market.Question,
			Probability: market.Probability,
		},
		Tags:              append([]string{"new", "market"}, content.Tags...),
		Significance:      models.SignificanceMedium,
		Sentiment:         content.Sentiment,
		MetaTitle:         content.Headline + " | FutureSignals",
		MetaDescription:   content.Summary,
		Published:         true,
		EnrichmentSources: sources,
	}

	if err := g.store.SaveArticle(ctx, article); err != nil {
		return nil, fmt.Errorf("failed to save article: %w", err)
	}

	log.Info().
		Str("slug", article.Slug).
		Msg("New market article generated")

	return article, nil
}

// GenerateCategoryDigest generates a digest for a specific category.
func (g *Generator) GenerateCategoryDigest(ctx context.Context, category string, limit int) (*models.Article, error) {
	log.Info().
		Str("category", category).
		Msg("Generating category digest")

	// Get markets for category
	markets, err := g.store.GetMarketsByCategory(ctx, category, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get markets: %w", err)
	}

	if len(markets) == 0 {
		return nil, fmt.Errorf("no markets found for category %s", category)
	}

	// Convert to refs
	var marketRefs []models.MarketRef
	for _, m := range markets {
		marketRefs = append(marketRefs, models.MarketRef{
			MarketID:    m.MarketID,
			Question:    m.Question,
			Slug:        m.Slug,
			Probability: m.Probability,
			Change24h:   m.Change24h,
			Volume24h:   m.Volume24h,
		})
	}

	// Generate content
	content, err := g.generateCategoryDigestContent(ctx, category, marketRefs)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	now := time.Now()
	catInfo := models.GetCategoryBySlug(category)
	catName := category
	if catInfo != nil {
		catName = catInfo.Name
	}

	slug := fmt.Sprintf("%s-digest-%s", category, now.Format("2006-01-02"))

	article := &models.Article{
		Slug:        slug,
		Type:        models.ArticleTypeDigest,
		Category:    category,
		Headline:    fmt.Sprintf("%s Markets: %s", catName, content.Headline),
		Subheadline: content.Summary,
		Summary:     content.Summary,
		Body: models.ArticleBody{
			WhatHappened: content.Overview,
			WhyItMatters: content.Analysis,
			Context:      content.Highlights,
			WhatToWatch:  content.WhatToWatch,
		},
		Markets:         marketRefs,
		Tags:            append([]string{category, "digest", "analysis"}, content.Tags...),
		Significance:    models.SignificanceMedium,
		Sentiment:       content.Sentiment,
		MetaTitle:       fmt.Sprintf("%s Prediction Markets Digest | FutureSignals", catName),
		MetaDescription: content.Summary,
		Published:       true,
	}

	if err := g.store.SaveArticle(ctx, article); err != nil {
		return nil, fmt.Errorf("failed to save article: %w", err)
	}

	log.Info().
		Str("slug", article.Slug).
		Int("markets", len(marketRefs)).
		Msg("Category digest generated")

	return article, nil
}

// Helper methods

func (g *Generator) generateSlug(headline string) string {
	slug := strings.ToLower(headline)
	slug = strings.ReplaceAll(slug, " ", "-")

	replacer := strings.NewReplacer(
		"'", "", "\"", "", "?", "", "!", "", ",", "", ".", "",
		":", "", ";", "", "(", "", ")", "",
	)
	slug = replacer.Replace(slug)

	if len(slug) > 80 {
		slug = slug[:80]
	}

	slug = strings.TrimRight(slug, "-")
	return slug + "-" + time.Now().Format("20060102-1504")
}

func (g *Generator) generateNarrative(ctx context.Context, market *models.Market, enrichedCtx, contentType string) (*qwen.Narrative, error) {
	if g.llm == nil {
		return nil, fmt.Errorf("LLM client not configured")
	}

	return g.llm.GenerateNarrative(ctx, qwen.SignalData{
		MarketTitle:     market.Question,
		EventTitle:      market.GroupItemTitle,
		Category:        market.Category,
		PreviousProb:    market.PreviousProb,
		CurrentProb:     market.Probability,
		TimeFrame:       "24h",
		Volume24h:       market.Volume24h,
		TotalVolume:     market.TotalVolume,
		ExternalContext: enrichedCtx,
	})
}

// LLM content types for different article types

type BriefingContent struct {
	Summary     string
	Overview    string
	KeyInsights string
	Highlights  []string
	WhatToWatch string
}

type TrendingContent struct {
	Headline    string
	Summary     string
	Overview    string
	Analysis    string
	Highlights  []string
	WhatToWatch string
	Tags        []string
}

type NewMarketContent struct {
	Headline    string
	Summary     string
	Overview    string
	WhyItMatters string
	Context     []string
	WhatToWatch string
	Tags        []string
	Sentiment   string
}

type CategoryDigestContent struct {
	Headline    string
	Summary     string
	Overview    string
	Analysis    string
	Highlights  []string
	WhatToWatch string
	Tags        []string
	Sentiment   string
}

func (g *Generator) generateBriefingContent(ctx context.Context, briefingType models.BriefingType, markets []models.MarketRef) (*BriefingContent, error) {
	if g.llm == nil {
		// Return basic content without LLM
		return &BriefingContent{
			Summary:     fmt.Sprintf("Your %s prediction market briefing with %d markets", briefingType, len(markets)),
			Overview:    "Here are the top prediction markets to watch.",
			KeyInsights: "Market activity continues across multiple categories.",
			Highlights:  []string{"Multiple high-volume markets active", "Prices moving across categories"},
			WhatToWatch: "Monitor these markets for significant movements.",
		}, nil
	}

	// Build market summary for prompt
	var marketSummary strings.Builder
	for i, m := range markets {
		if i >= 10 { // Limit to 10 for prompt
			break
		}
		marketSummary.WriteString(fmt.Sprintf("- %s: %.1f%% (%.1f%% change, $%.0f volume)\n",
			m.Question, m.Probability*100, m.Change24h*100, m.Volume24h))
	}

	prompt := fmt.Sprintf(`Generate a %s briefing for prediction markets.

MARKETS:
%s

Generate JSON:
{
  "summary": "2-sentence executive summary",
  "overview": "3-4 sentences covering the main themes",
  "key_insights": "2-3 key insights from the data",
  "highlights": ["highlight 1", "highlight 2", "highlight 3"],
  "what_to_watch": "1-2 sentences on what to monitor"
}`, briefingType, marketSummary.String())

	var result struct {
		Summary     string   `json:"summary"`
		Overview    string   `json:"overview"`
		KeyInsights string   `json:"key_insights"`
		Highlights  []string `json:"highlights"`
		WhatToWatch string   `json:"what_to_watch"`
	}

	err := g.llm.ChatJSON(ctx, qwen.ChatRequest{
		SystemPrompt: "You are a financial markets analyst. Generate concise, professional market briefings.",
		UserPrompt:   prompt,
		Temperature:  0.3,
		MaxTokens:    800,
	}, &result)

	if err != nil {
		return nil, err
	}

	return &BriefingContent{
		Summary:     result.Summary,
		Overview:    result.Overview,
		KeyInsights: result.KeyInsights,
		Highlights:  result.Highlights,
		WhatToWatch: result.WhatToWatch,
	}, nil
}

func (g *Generator) generateTrendingContent(ctx context.Context, markets []models.MarketRef) (*TrendingContent, error) {
	if g.llm == nil {
		return &TrendingContent{
			Headline:    fmt.Sprintf("Top %d Trending Prediction Markets", len(markets)),
			Summary:     "The hottest prediction markets right now based on volume and activity.",
			Overview:    "These markets are seeing the most trading activity.",
			Analysis:    "High volume indicates strong trader interest.",
			Highlights:  []string{"Multiple markets showing elevated activity"},
			WhatToWatch: "Monitor for continued momentum.",
			Tags:        []string{},
		}, nil
	}

	var marketSummary strings.Builder
	for i, m := range markets {
		if i >= 10 {
			break
		}
		marketSummary.WriteString(fmt.Sprintf("- %s: %.1f%% ($%.0fK volume)\n",
			m.Question, m.Probability*100, m.Volume24h/1000))
	}

	prompt := fmt.Sprintf(`Analyze these trending prediction markets:

%s

Generate JSON:
{
  "headline": "Compelling headline (max 80 chars)",
  "summary": "2-sentence summary",
  "overview": "3-4 sentences on what's trending",
  "analysis": "2-3 sentences on why these are hot",
  "highlights": ["key point 1", "key point 2"],
  "what_to_watch": "What to monitor next",
  "tags": ["relevant", "tags"]
}`, marketSummary.String())

	var result TrendingContent
	err := g.llm.ChatJSON(ctx, qwen.ChatRequest{
		SystemPrompt: "You are a financial markets analyst covering prediction markets.",
		UserPrompt:   prompt,
		Temperature:  0.3,
		MaxTokens:    600,
	}, &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (g *Generator) generateNewMarketContent(ctx context.Context, market *models.Market, enrichedCtx string) (*NewMarketContent, error) {
	if g.llm == nil {
		return &NewMarketContent{
			Headline:     fmt.Sprintf("New Market: %s", truncate(market.Question, 60)),
			Summary:      fmt.Sprintf("A new prediction market asks: %s", market.Question),
			Overview:     "This market has just been created and is now accepting trades.",
			WhyItMatters: "New markets offer opportunities to express views on emerging topics.",
			Context:      []string{},
			WhatToWatch:  "Watch for early price discovery and volume.",
			Tags:         []string{market.Category},
			Sentiment:    "neutral",
		}, nil
	}

	prompt := fmt.Sprintf(`A new prediction market was just created:

QUESTION: %s
CATEGORY: %s
CURRENT PROBABILITY: %.1f%%
INITIAL VOLUME: $%.0f

CONTEXT (if available):
%s

Generate JSON:
{
  "headline": "Compelling headline about this new market",
  "summary": "2-sentence summary",
  "overview": "What this market is about",
  "why_it_matters": "Why traders should care",
  "context": ["relevant context point 1", "point 2"],
  "what_to_watch": "What could move this market",
  "tags": ["relevant", "tags"],
  "sentiment": "bullish|bearish|neutral"
}`, market.Question, market.Category, market.Probability*100, market.Volume24h, enrichedCtx)

	var result NewMarketContent
	err := g.llm.ChatJSON(ctx, qwen.ChatRequest{
		SystemPrompt: "You are a financial journalist covering prediction markets.",
		UserPrompt:   prompt,
		Temperature:  0.4,
		MaxTokens:    600,
	}, &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (g *Generator) generateCategoryDigestContent(ctx context.Context, category string, markets []models.MarketRef) (*CategoryDigestContent, error) {
	catInfo := models.GetCategoryBySlug(category)
	catName := category
	if catInfo != nil {
		catName = catInfo.Name
	}

	if g.llm == nil {
		return &CategoryDigestContent{
			Headline:    fmt.Sprintf("What's Moving in %s", catName),
			Summary:     fmt.Sprintf("A look at the top %s prediction markets.", catName),
			Overview:    fmt.Sprintf("Here are the most active %s markets.", catName),
			Analysis:    "Market activity reflects current events and sentiment.",
			Highlights:  []string{},
			WhatToWatch: "Monitor for significant movements.",
			Tags:        []string{},
			Sentiment:   "neutral",
		}, nil
	}

	var marketSummary strings.Builder
	for i, m := range markets {
		if i >= 10 {
			break
		}
		marketSummary.WriteString(fmt.Sprintf("- %s: %.1f%% (%.1f%% change)\n",
			m.Question, m.Probability*100, m.Change24h*100))
	}

	prompt := fmt.Sprintf(`Create a digest for %s prediction markets:

MARKETS:
%s

Generate JSON:
{
  "headline": "Compelling digest headline",
  "summary": "2-sentence executive summary",
  "overview": "3-4 sentences on category state",
  "analysis": "Key insights and patterns",
  "highlights": ["key point 1", "key point 2"],
  "what_to_watch": "What to monitor",
  "tags": ["relevant", "tags"],
  "sentiment": "bullish|bearish|neutral"
}`, catName, marketSummary.String())

	var result CategoryDigestContent
	err := g.llm.ChatJSON(ctx, qwen.ChatRequest{
		SystemPrompt: "You are a financial analyst specializing in prediction markets.",
		UserPrompt:   prompt,
		Temperature:  0.3,
		MaxTokens:    600,
	}, &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

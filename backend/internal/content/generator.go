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
		return &BriefingContent{
			Summary:     fmt.Sprintf("Your %s prediction market briefing with %d markets", briefingType, len(markets)),
			Overview:    "Here are the top prediction markets to watch.",
			KeyInsights: "Market activity continues across multiple categories.",
			Highlights:  []string{"Multiple high-volume markets active", "Prices moving across categories"},
			WhatToWatch: "Monitor these markets for significant movements.",
		}, nil
	}

	// Build market summary with Bloomberg-style data integration
	var marketSummary strings.Builder
	totalVolume := 0.0
	biggestMover := ""
	biggestMove := 0.0

	for i, m := range markets {
		if i >= 10 {
			break
		}
		totalVolume += m.Volume24h
		if abs(m.Change24h) > abs(biggestMove) {
			biggestMove = m.Change24h
			biggestMover = m.Question
		}
		marketSummary.WriteString(fmt.Sprintf("• %s: %.0f%% (%+.1fpts, $%.0fK vol)\n",
			m.Question, m.Probability*100, m.Change24h*100, m.Volume24h/1000))
	}

	systemPrompt := `You are a senior financial journalist writing a market briefing in Bloomberg wire service style.

STYLE GUIDE:
- Lead with the most significant development
- Integrate specific numbers into prose (not bullet points in the output)
- Short, punchy sentences
- Answer "so what?" for sophisticated readers
- Forward-looking closing

Respond ONLY with valid JSON.`

	prompt := fmt.Sprintf(`Write a %s MARKET BRIEFING in Bloomberg style.

═══════════════════════════════════════════════════════════════
MARKET DATA
═══════════════════════════════════════════════════════════════
Total 24h Volume: $%.1fM
Biggest Mover: %s (%+.1f points)

MARKETS:
%s

═══════════════════════════════════════════════════════════════
OUTPUT
═══════════════════════════════════════════════════════════════
{
  "summary": "Bloomberg-style 2-sentence executive summary. Lead with the biggest story. Include specific numbers.",
  "overview": "3-4 sentences covering main market themes. Weave in specific data. Explain what's driving activity.",
  "key_insights": "2-3 sentences of analysis. What patterns emerge? What do the odds imply? Connect to real-world events.",
  "highlights": ["Specific highlight with data", "Another concrete observation", "Forward-looking point"],
  "what_to_watch": "2 sentences on upcoming catalysts. Be specific about dates/events that could move markets."
}`, briefingType, totalVolume/1_000_000, biggestMover, biggestMove*100, marketSummary.String())

	var result struct {
		Summary     string   `json:"summary"`
		Overview    string   `json:"overview"`
		KeyInsights string   `json:"key_insights"`
		Highlights  []string `json:"highlights"`
		WhatToWatch string   `json:"what_to_watch"`
	}

	err := g.llm.ChatJSON(ctx, qwen.ChatRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   prompt,
		Temperature:  0.4,
		MaxTokens:    1000,
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

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
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

	// Calculate aggregate stats
	var marketSummary strings.Builder
	totalVolume := 0.0
	topMarket := ""
	topVolume := 0.0

	for i, m := range markets {
		if i >= 10 {
			break
		}
		totalVolume += m.Volume24h
		if m.Volume24h > topVolume {
			topVolume = m.Volume24h
			topMarket = m.Question
		}
		marketSummary.WriteString(fmt.Sprintf("• %s: %.0f%% ($%.0fK 24h vol, %+.1fpts)\n",
			m.Question, m.Probability*100, m.Volume24h/1000, m.Change24h*100))
	}

	systemPrompt := `You are a senior financial journalist at a wire service covering prediction markets.

STYLE: Bloomberg/Reuters wire service
- Active voice headlines with specific numbers
- Lead with the most newsworthy angle
- Integrate data into narrative prose
- Answer "why is this trending?" and "so what?"
- Short, punchy sentences

Respond ONLY with valid JSON.`

	prompt := fmt.Sprintf(`Write a TRENDING MARKETS story in Bloomberg wire style.

═══════════════════════════════════════════════════════════════
AGGREGATE DATA
═══════════════════════════════════════════════════════════════
Combined 24h Volume: $%.1fM
Top Volume Market: %s ($%.0fK)

TRENDING MARKETS:
%s

═══════════════════════════════════════════════════════════════
OUTPUT
═══════════════════════════════════════════════════════════════
{
  "headline": "Active-voice headline with key number. Max 80 chars. Example: 'Prediction Markets See $5M Flow Into Election Bets'",
  "summary": "2-sentence wire-style summary. Lead with the biggest story, include specific volume/probability figures.",
  "overview": "3-4 sentences explaining what's driving volume. Connect to real-world events. Why are traders active now?",
  "analysis": "2-3 sentences of market analysis. What do the odds imply? What's the smart money saying?",
  "highlights": ["Specific observation with data", "Pattern or trend identified", "Forward-looking point"],
  "what_to_watch": "2 sentences on upcoming catalysts that could drive more activity.",
  "tags": ["relevant", "seo", "tags"]
}`, totalVolume/1_000_000, topMarket, topVolume/1000, marketSummary.String())

	var result TrendingContent
	err := g.llm.ChatJSON(ctx, qwen.ChatRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   prompt,
		Temperature:  0.4,
		MaxTokens:    800,
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

	// Determine implied odds interpretation
	impliedOutcome := "uncertain"
	if market.Probability > 0.7 {
		impliedOutcome = "likely"
	} else if market.Probability < 0.3 {
		impliedOutcome = "unlikely"
	}

	systemPrompt := `You are a senior financial journalist covering new prediction market listings.

STYLE: Bloomberg/Reuters wire service
- Explain the market's significance in broader context
- Connect to current events when possible
- Integrate the probability data into narrative
- Short, punchy sentences
- Answer "why should readers care about this new market?"

Respond ONLY with valid JSON.`

	contextStr := enrichedCtx
	if contextStr == "" {
		contextStr = "No additional context available."
	}

	prompt := fmt.Sprintf(`Write a NEW MARKET LISTING story in Bloomberg wire style.

═══════════════════════════════════════════════════════════════
NEW MARKET
═══════════════════════════════════════════════════════════════
Question: %s
Category: %s
Opening Probability: %.0f%% (implied: %s)
Initial Volume: $%.0fK
End Date: %s

External Context:
%s

═══════════════════════════════════════════════════════════════
OUTPUT
═══════════════════════════════════════════════════════════════
{
  "headline": "Active-voice headline announcing the market. Include the opening odds. Max 80 chars.",
  "summary": "2-sentence wire-style summary. What is the market, what are opening odds, why now?",
  "overview": "2-3 sentences explaining the market and its context. Connect to real-world events or decisions.",
  "why_it_matters": "2-3 sentences on stakes. What happens if this resolves Yes/No? Economic/political implications?",
  "context": ["Relevant background fact with data", "Another contextual point"],
  "what_to_watch": "2 sentences on what could move this market. Key dates, events, catalysts.",
  "tags": ["relevant", "seo", "tags"],
  "sentiment": "bullish|bearish|neutral"
}`, market.Question, market.Category, market.Probability*100, impliedOutcome, market.Volume24h/1000, market.EndDate, contextStr)

	var result NewMarketContent
	err := g.llm.ChatJSON(ctx, qwen.ChatRequest{
		SystemPrompt: systemPrompt,
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

	// Build market summary with aggregate stats
	var marketSummary strings.Builder
	totalVolume := 0.0
	avgProb := 0.0
	bullishCount := 0
	bearishCount := 0

	for i, m := range markets {
		if i >= 10 {
			break
		}
		totalVolume += m.Volume24h
		avgProb += m.Probability
		if m.Change24h > 0.02 {
			bullishCount++
		} else if m.Change24h < -0.02 {
			bearishCount++
		}
		marketSummary.WriteString(fmt.Sprintf("• %s: %.0f%% (%+.1fpts, $%.0fK vol)\n",
			m.Question, m.Probability*100, m.Change24h*100, m.Volume24h/1000))
	}

	marketCount := len(markets)
	if marketCount > 10 {
		marketCount = 10
	}
	if marketCount > 0 {
		avgProb /= float64(marketCount)
	}

	// Determine overall sentiment
	overallSentiment := "mixed"
	if bullishCount > bearishCount*2 {
		overallSentiment = "bullish"
	} else if bearishCount > bullishCount*2 {
		overallSentiment = "bearish"
	}

	systemPrompt := `You are a senior financial journalist writing a sector digest in Bloomberg wire service style.

STYLE:
- Lead with the most significant development in this category
- Integrate specific numbers into prose
- Connect market movements to real-world events
- Explain what the odds imply for the category
- Short, authoritative sentences

Respond ONLY with valid JSON.`

	prompt := fmt.Sprintf(`Write a %s CATEGORY DIGEST in Bloomberg wire style.

═══════════════════════════════════════════════════════════════
CATEGORY STATS
═══════════════════════════════════════════════════════════════
Category: %s
Combined 24h Volume: $%.1fM
Average Probability: %.0f%%
Sentiment: %d bullish / %d bearish moves
Overall Trend: %s

MARKETS:
%s

═══════════════════════════════════════════════════════════════
OUTPUT
═══════════════════════════════════════════════════════════════
{
  "headline": "Active-voice headline capturing category story. Include key data. Max 80 chars.",
  "summary": "2-sentence wire-style summary. Lead with the biggest story in this category.",
  "overview": "3-4 sentences on category state. What themes are dominating? Connect to real events.",
  "analysis": "2-3 sentences of analysis. What do the collective odds suggest? Any patterns?",
  "highlights": ["Specific highlight with data", "Pattern or trend", "Forward-looking point"],
  "what_to_watch": "2 sentences on upcoming catalysts for this category.",
  "tags": ["relevant", "seo", "tags"],
  "sentiment": "bullish|bearish|neutral"
}`, catName, catName, totalVolume/1_000_000, avgProb*100, bullishCount, bearishCount, overallSentiment, marketSummary.String())

	var result CategoryDigestContent
	err := g.llm.ChatJSON(ctx, qwen.ChatRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   prompt,
		Temperature:  0.4,
		MaxTokens:    1000,
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

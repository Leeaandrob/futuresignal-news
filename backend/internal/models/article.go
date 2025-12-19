package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ArticleType represents the type of article.
type ArticleType string

const (
	// ArticleTypeBreaking represents breaking news from significant market movements.
	ArticleTypeBreaking ArticleType = "breaking"

	// ArticleTypeBriefing represents scheduled briefings (morning, midday, evening).
	ArticleTypeBriefing ArticleType = "briefing"

	// ArticleTypeTrending represents trending market analysis.
	ArticleTypeTrending ArticleType = "trending"

	// ArticleTypeNewMarket represents coverage of new markets.
	ArticleTypeNewMarket ArticleType = "new_market"

	// ArticleTypeDeepDive represents in-depth analysis of a single market.
	ArticleTypeDeepDive ArticleType = "deep_dive"

	// ArticleTypeDigest represents category or weekly digests.
	ArticleTypeDigest ArticleType = "digest"

	// ArticleTypeExplainer represents educational content.
	ArticleTypeExplainer ArticleType = "explainer"
)

// Significance represents the importance level of an article.
type Significance string

const (
	SignificanceLow      Significance = "low"
	SignificanceMedium   Significance = "medium"
	SignificanceHigh     Significance = "high"
	SignificanceBreaking Significance = "breaking"
)

// Article represents a generated article/news piece.
type Article struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	// Identifiers
	Slug string `bson:"slug" json:"slug"`

	// Classification
	Type     ArticleType `bson:"type" json:"type"`
	Category string      `bson:"category" json:"category"`

	// Content
	Headline    string      `bson:"headline" json:"headline"`
	Subheadline string      `bson:"subheadline" json:"subheadline"`
	Summary     string      `bson:"summary" json:"summary"`
	Body        ArticleBody `bson:"body" json:"body"`

	// Related Markets
	Markets       []MarketRef `bson:"markets" json:"markets"`
	PrimaryMarket *MarketRef  `bson:"primary_market,omitempty" json:"primary_market,omitempty"`

	// Metadata
	Tags         []string     `bson:"tags" json:"tags"`
	Significance Significance `bson:"significance" json:"significance"`
	Sentiment    string       `bson:"sentiment" json:"sentiment"` // bullish, bearish, neutral

	// Timing
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	PublishedAt time.Time `bson:"published_at" json:"published_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`

	// SEO
	MetaTitle       string `bson:"meta_title" json:"meta_title"`
	MetaDescription string `bson:"meta_description" json:"meta_description"`
	CanonicalURL    string `bson:"canonical_url,omitempty" json:"canonical_url,omitempty"`

	// Stats
	Views int `bson:"views" json:"views"`

	// Status
	Published bool `bson:"published" json:"published"`
	Featured  bool `bson:"featured" json:"featured"`

	// Enrichment sources used
	EnrichmentSources []string `bson:"enrichment_sources,omitempty" json:"enrichment_sources,omitempty"`
}

// ArticleBody contains the main content sections.
type ArticleBody struct {
	WhatHappened string   `bson:"what_happened" json:"what_happened"`
	WhyItMatters string   `bson:"why_it_matters" json:"why_it_matters"`
	Context      []string `bson:"context" json:"context"`
	WhatToWatch  string   `bson:"what_to_watch" json:"what_to_watch"`
	Analysis     string   `bson:"analysis,omitempty" json:"analysis,omitempty"`
}

// MarketRef references a market within an article.
type MarketRef struct {
	MarketID      string  `bson:"market_id" json:"market_id"`
	Question      string  `bson:"question" json:"question"`
	Slug          string  `bson:"slug" json:"slug"`
	Probability   float64 `bson:"probability" json:"probability"`
	PreviousProb  float64 `bson:"previous_prob,omitempty" json:"previous_prob,omitempty"`
	Change24h     float64 `bson:"change_24h" json:"change_24h"`
	Volume24h     float64 `bson:"volume_24h" json:"volume_24h"`
	TotalVolume   float64 `bson:"total_volume" json:"total_volume"`
	EndDate       string  `bson:"end_date,omitempty" json:"end_date,omitempty"`
}

// BriefingType represents the type of scheduled briefing.
type BriefingType string

const (
	BriefingMorning BriefingType = "morning"
	BriefingMidday  BriefingType = "midday"
	BriefingEvening BriefingType = "evening"
	BriefingWeekly  BriefingType = "weekly"
)

// BriefingConfig holds configuration for briefing generation.
type BriefingConfig struct {
	Type           BriefingType
	Title          string
	MarketsPerCat  int
	Categories     []string
	IncludeSummary bool
}

// DefaultBriefingConfigs returns the default briefing configurations.
var DefaultBriefingConfigs = map[BriefingType]BriefingConfig{
	BriefingMorning: {
		Type:           BriefingMorning,
		Title:          "Morning Market Briefing",
		MarketsPerCat:  3,
		Categories:     []string{"politics", "crypto", "finance", "tech", "sports"},
		IncludeSummary: true,
	},
	BriefingMidday: {
		Type:           BriefingMidday,
		Title:          "Midday Market Pulse",
		MarketsPerCat:  2,
		Categories:     []string{"politics", "crypto", "finance"},
		IncludeSummary: false,
	},
	BriefingEvening: {
		Type:           BriefingEvening,
		Title:          "Evening Market Wrap",
		MarketsPerCat:  3,
		Categories:     []string{"politics", "crypto", "finance", "tech", "sports"},
		IncludeSummary: true,
	},
	BriefingWeekly: {
		Type:           BriefingWeekly,
		Title:          "Weekly Market Digest",
		MarketsPerCat:  5,
		Categories:     []string{"politics", "crypto", "finance", "tech", "sports", "geopolitics"},
		IncludeSummary: true,
	},
}

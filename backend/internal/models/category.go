// Package models defines the core data structures for FutureSignals.
package models

// Category represents a content category.
type Category struct {
	ID          string `bson:"_id" json:"id"`
	Slug        string `bson:"slug" json:"slug"`
	Name        string `bson:"name" json:"name"`
	Description string `bson:"description" json:"description"`
	Icon        string `bson:"icon" json:"icon"`
	Color       string `bson:"color" json:"color"`
	Order       int    `bson:"order" json:"order"`
	Dynamic     bool   `bson:"dynamic" json:"dynamic"` // trending, breaking, new are dynamic
}

// DefaultCategories mirrors Polymarket's category structure.
var DefaultCategories = []Category{
	// Dynamic categories (computed, not assigned)
	{Slug: "trending", Name: "Trending", Description: "Most active prediction markets right now", Icon: "trending_up", Color: "#FF6B6B", Order: 1, Dynamic: true},
	{Slug: "breaking", Name: "Breaking", Description: "Significant market movements and news", Icon: "bolt", Color: "#FF4757", Order: 2, Dynamic: true},
	{Slug: "new", Name: "New", Description: "Recently created markets", Icon: "fiber_new", Color: "#2ED573", Order: 3, Dynamic: true},

	// Static categories (assigned to markets)
	{Slug: "politics", Name: "Politics", Description: "Political predictions and elections", Icon: "account_balance", Color: "#5352ED", Order: 10},
	{Slug: "elections", Name: "Elections", Description: "Election predictions worldwide", Icon: "how_to_vote", Color: "#A29BFE", Order: 11},
	{Slug: "crypto", Name: "Crypto", Description: "Cryptocurrency predictions", Icon: "currency_bitcoin", Color: "#F7931A", Order: 20},
	{Slug: "finance", Name: "Finance", Description: "Financial markets and economic predictions", Icon: "trending_up", Color: "#00D2D3", Order: 21},
	{Slug: "economy", Name: "Economy", Description: "Economic indicators and predictions", Icon: "payments", Color: "#FDCB6E", Order: 22},
	{Slug: "earnings", Name: "Earnings", Description: "Company earnings predictions", Icon: "attach_money", Color: "#00B894", Order: 23},
	{Slug: "tech", Name: "Tech", Description: "Technology industry predictions", Icon: "computer", Color: "#0984E3", Order: 30},
	{Slug: "sports", Name: "Sports", Description: "Sports predictions and outcomes", Icon: "sports_soccer", Color: "#1E90FF", Order: 40},
	{Slug: "geopolitics", Name: "Geopolitics", Description: "Global political events and conflicts", Icon: "public", Color: "#6C5CE7", Order: 50},
	{Slug: "world", Name: "World", Description: "Global events and news", Icon: "language", Color: "#636E72", Order: 51},
	{Slug: "culture", Name: "Culture", Description: "Pop culture and entertainment", Icon: "movie", Color: "#E84393", Order: 60},
}

// CategoryKeywords maps keywords to categories for auto-detection.
var CategoryKeywords = map[string][]string{
	"politics": {
		"president", "congress", "senate", "house", "vote", "trump", "biden",
		"government", "governor", "mayor", "legislation", "bill", "law",
		"republican", "democrat", "gop", "dnc", "rnc", "white house",
	},
	"elections": {
		"election", "ballot", "primary", "nominee", "electoral", "swing state",
		"poll", "voter", "voting", "candidate", "midterm", "runoff",
	},
	"crypto": {
		"bitcoin", "btc", "ethereum", "eth", "crypto", "token", "blockchain",
		"defi", "nft", "altcoin", "stablecoin", "usdc", "usdt", "solana",
		"cardano", "dogecoin", "shiba", "binance", "coinbase", "sec crypto",
	},
	"finance": {
		"stock", "nasdaq", "dow", "s&p", "market", "trading", "investor",
		"wall street", "hedge fund", "ipo", "merger", "acquisition",
	},
	"economy": {
		"fed", "federal reserve", "interest rate", "inflation", "gdp",
		"recession", "unemployment", "jobs report", "cpi", "treasury",
		"fiscal", "monetary", "debt ceiling", "deficit",
	},
	"earnings": {
		"earnings", "revenue", "profit", "quarterly", "eps", "guidance",
		"beat", "miss", "forecast", "outlook",
	},
	"tech": {
		"ai", "artificial intelligence", "openai", "chatgpt", "google", "apple",
		"microsoft", "meta", "amazon", "tesla", "nvidia", "semiconductor",
		"chip", "software", "startup", "silicon valley", "spacex", "elon",
	},
	"sports": {
		"nfl", "nba", "mlb", "nhl", "soccer", "football", "basketball",
		"baseball", "hockey", "super bowl", "world series", "championship",
		"playoffs", "finals", "mvp", "draft", "trade", "coach",
	},
	"geopolitics": {
		"war", "conflict", "military", "nato", "russia", "ukraine", "china",
		"taiwan", "iran", "israel", "palestine", "ceasefire", "sanctions",
		"treaty", "summit", "diplomacy", "embassy",
	},
	"world": {
		"international", "global", "united nations", "un", "world",
		"foreign", "abroad", "overseas",
	},
	"culture": {
		"movie", "film", "oscars", "grammy", "emmys", "celebrity", "music",
		"album", "tour", "concert", "tv show", "streaming", "netflix",
		"disney", "marvel", "box office", "viral", "tiktok", "influencer",
	},
}

// GetCategoryBySlug returns a category by its slug.
func GetCategoryBySlug(slug string) *Category {
	for _, cat := range DefaultCategories {
		if cat.Slug == slug {
			return &cat
		}
	}
	return nil
}

// GetStaticCategories returns non-dynamic categories.
func GetStaticCategories() []Category {
	var static []Category
	for _, cat := range DefaultCategories {
		if !cat.Dynamic {
			static = append(static, cat)
		}
	}
	return static
}

// GetDynamicCategories returns dynamic categories.
func GetDynamicCategories() []Category {
	var dynamic []Category
	for _, cat := range DefaultCategories {
		if cat.Dynamic {
			dynamic = append(dynamic, cat)
		}
	}
	return dynamic
}

// CategorySentiment represents momentum/sentiment data for a category.
type CategorySentiment struct {
	Category       string  `bson:"category" json:"category"`
	Name           string  `bson:"name" json:"name"`
	Color          string  `bson:"color" json:"color"`
	Icon           string  `bson:"icon" json:"icon"`
	Momentum       float64 `bson:"momentum" json:"momentum"`               // Volume-weighted avg change (-1 to 1)
	TotalVolume24h float64 `bson:"total_volume_24h" json:"total_volume_24h"` // Sum of all volume24h
	MarketCount    int     `bson:"market_count" json:"market_count"`       // Active markets count
	BreakingCount  int     `bson:"breaking_count" json:"breaking_count"`   // Markets with |change| > 10%
	TopMover       string  `bson:"top_mover,omitempty" json:"top_mover,omitempty"`           // Market with highest |change|
	TopMoverSlug   string  `bson:"top_mover_slug,omitempty" json:"top_mover_slug,omitempty"` // Slug for link
	TopMoverChange float64 `bson:"top_mover_change" json:"top_mover_change"`                 // Change of top mover
	AvgChange24h   float64 `bson:"avg_change_24h" json:"avg_change_24h"`   // Simple average change
}

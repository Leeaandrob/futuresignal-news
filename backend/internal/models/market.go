package models

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Market represents a prediction market from Polymarket.
type Market struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	// Polymarket identifiers
	MarketID       string `bson:"market_id" json:"market_id"`
	ConditionID    string `bson:"condition_id" json:"condition_id"`
	Slug           string `bson:"slug" json:"slug"`
	GroupItemTitle string `bson:"group_item_title,omitempty" json:"group_item_title,omitempty"`

	// Content
	Question    string `bson:"question" json:"question"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`

	// Classification
	Category string   `bson:"category" json:"category"`
	Tags     []string `bson:"tags" json:"tags"`

	// Market data
	Probability  float64 `bson:"probability" json:"probability"` // Current yes price
	PreviousProb float64 `bson:"previous_prob" json:"previous_prob"`
	Change1h     float64 `bson:"change_1h" json:"change_1h"`
	Change24h    float64 `bson:"change_24h" json:"change_24h"`
	Change7d     float64 `bson:"change_7d" json:"change_7d"`

	// Volume
	Volume1h    float64 `bson:"volume_1h" json:"volume_1h"`
	Volume24h   float64 `bson:"volume_24h" json:"volume_24h"`
	Volume7d    float64 `bson:"volume_7d" json:"volume_7d"`
	TotalVolume float64 `bson:"total_volume" json:"total_volume"`

	// Liquidity
	Liquidity float64 `bson:"liquidity" json:"liquidity"`

	// Status
	Active       bool   `bson:"active" json:"active"`
	Closed       bool   `bson:"closed" json:"closed"`
	Archived     bool   `bson:"archived" json:"archived"`
	AcceptingBid bool   `bson:"accepting_bid" json:"accepting_bid"`
	EndDate      string `bson:"end_date,omitempty" json:"end_date,omitempty"`

	// Outcomes (for multi-outcome markets)
	Outcomes      []string  `bson:"outcomes" json:"outcomes"`
	OutcomePrices []float64 `bson:"outcome_prices" json:"outcome_prices"`

	// Timing
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
	FirstSeenAt time.Time `bson:"first_seen_at" json:"first_seen_at"`

	// Trending score (calculated)
	TrendingScore float64 `bson:"trending_score" json:"trending_score"`

	// URL
	PolymarketURL string `bson:"polymarket_url" json:"polymarket_url"`
}

// Snapshot represents a historical snapshot of market data.
type Snapshot struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	MarketID    string    `bson:"market_id" json:"market_id"`
	Probability float64   `bson:"probability" json:"probability"`
	Volume24h   float64   `bson:"volume_24h" json:"volume_24h"`
	TotalVolume float64   `bson:"total_volume" json:"total_volume"`
	Liquidity   float64   `bson:"liquidity" json:"liquidity"`
	CapturedAt  time.Time `bson:"captured_at" json:"captured_at"`
}

// TrendingMetrics holds data for trending calculation.
type TrendingMetrics struct {
	VolumeScore    float64 // Based on recent volume
	MovementScore  float64 // Based on price movement
	VelocityScore  float64 // Based on rate of change
	RecencyScore   float64 // Based on how recent the activity is
	TotalScore     float64 // Combined score
}

// CalculateTrendingScore calculates a trending score for the market.
func (m *Market) CalculateTrendingScore() float64 {
	// Volume component (0-40 points)
	volumeScore := 0.0
	switch {
	case m.Volume24h >= 1000000:
		volumeScore = 40
	case m.Volume24h >= 500000:
		volumeScore = 30
	case m.Volume24h >= 100000:
		volumeScore = 20
	case m.Volume24h >= 50000:
		volumeScore = 10
	}

	// Movement component (0-30 points)
	movementScore := 0.0
	absChange := abs(m.Change24h)
	switch {
	case absChange >= 0.15:
		movementScore = 30
	case absChange >= 0.10:
		movementScore = 25
	case absChange >= 0.05:
		movementScore = 15
	case absChange >= 0.02:
		movementScore = 10
	}

	// Velocity component - hourly vs daily (0-20 points)
	velocityScore := 0.0
	if m.Volume24h > 0 && m.Volume1h > 0 {
		hourlyRatio := m.Volume1h / (m.Volume24h / 24)
		switch {
		case hourlyRatio >= 5:
			velocityScore = 20
		case hourlyRatio >= 3:
			velocityScore = 15
		case hourlyRatio >= 2:
			velocityScore = 10
		}
	}

	// Probability interest (0-10 points) - markets near 50% are more interesting
	interestScore := 10 - abs(m.Probability-0.5)*20

	return volumeScore + movementScore + velocityScore + interestScore
}

// DetectCategory attempts to categorize the market based on its question.
func (m *Market) DetectCategory() string {
	questionLower := strings.ToLower(m.Question)

	for category, keywords := range CategoryKeywords {
		for _, keyword := range keywords {
			if strings.Contains(questionLower, keyword) {
				return category
			}
		}
	}

	return "other"
}

// IsNew returns true if the market was first seen within the given duration.
func (m *Market) IsNew(within time.Duration) bool {
	return time.Since(m.FirstSeenAt) <= within
}

// IsBreaking returns true if the market has significant recent movement.
func (m *Market) IsBreaking(threshold float64) bool {
	return abs(m.Change24h) >= threshold
}

// IsTrending returns true if the market's trending score is above threshold.
func (m *Market) IsTrending(threshold float64) bool {
	return m.TrendingScore >= threshold
}

// Helper
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// GenerateSlug creates a URL-friendly slug from the question.
func (m *Market) GenerateSlug() string {
	slug := strings.ToLower(m.Question)
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters
	replacer := strings.NewReplacer(
		"'", "", "\"", "", "?", "", "!", "", ",", "", ".", "",
		":", "", ";", "", "(", "", ")", "", "[", "", "]", "",
		"&", "and", "%", "percent", "$", "usd", "@", "at",
	)
	slug = replacer.Replace(slug)

	// Truncate if too long
	if len(slug) > 80 {
		slug = slug[:80]
	}

	// Remove trailing hyphens
	slug = strings.TrimRight(slug, "-")

	return slug
}

package xtracker

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/leeaandrob/futuresignals/internal/models"
	"github.com/leeaandrob/futuresignals/internal/storage"
	"github.com/rs/zerolog/log"
)

// CorrelationConfig holds configuration for signal correlation.
type CorrelationConfig struct {
	// TimeWindow is how long after a tweet to look for market movements.
	TimeWindow time.Duration
	// MinMarketChange is the minimum % change to consider (e.g., 0.02 = 2%).
	MinMarketChange float64
	// MaxSignalsPerArticle limits signals added to an article.
	MaxSignalsPerArticle int
	// Categories to check for correlations.
	Categories []string
}

// DefaultCorrelationConfig returns sensible defaults.
func DefaultCorrelationConfig() CorrelationConfig {
	return CorrelationConfig{
		TimeWindow:           2 * time.Hour,
		MinMarketChange:      0.02, // 2%
		MaxSignalsPerArticle: 3,
		Categories:           []string{"politics", "tech", "crypto", "finance", "world"},
	}
}

// Correlator finds relationships between social signals and market movements.
type Correlator struct {
	client *Client
	store  *storage.Store
	config CorrelationConfig

	// Cache of tracked users
	users    []TrackedUser
	usersAt  time.Time
	cacheTTL time.Duration
}

// NewCorrelator creates a new signal correlator.
func NewCorrelator(client *Client, store *storage.Store, config CorrelationConfig) *Correlator {
	return &Correlator{
		client:   client,
		store:    store,
		config:   config,
		cacheTTL: 5 * time.Minute,
	}
}

// GetTrackedUsers returns cached tracked users or fetches fresh data.
func (c *Correlator) GetTrackedUsers(ctx context.Context) ([]TrackedUser, error) {
	if time.Since(c.usersAt) < c.cacheTTL && len(c.users) > 0 {
		return c.users, nil
	}

	users, err := c.client.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	c.users = users
	c.usersAt = time.Now()
	return users, nil
}

// FindSignalsForMarket finds social signals that may have influenced a market movement.
func (c *Correlator) FindSignalsForMarket(ctx context.Context, market *models.Market, lookback time.Duration) ([]models.SocialSignal, error) {
	users, err := c.GetTrackedUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting tracked users: %w", err)
	}

	var signals []models.SocialSignal
	since := time.Now().Add(-lookback)

	for _, user := range users {
		posts, err := c.client.GetRecentPosts(ctx, user.Handle, since, 50)
		if err != nil {
			log.Warn().Err(err).Str("handle", user.Handle).Msg("Failed to get posts")
			continue
		}

		for _, post := range posts {
			// Check if post content is relevant to the market
			if !c.isRelevant(post.Content, market) {
				continue
			}

			// Check timing correlation
			if !c.isTimeCorrelated(post.CreatedAt, market) {
				continue
			}

			signal := models.SocialSignal{
				Handle:       user.Handle,
				Name:         user.Name,
				AvatarURL:    user.AvatarURL,
				Verified:     user.Verified,
				Content:      truncateContent(post.Content, 280),
				TweetURL:     post.TweetURL(user.Handle),
				PostedAt:     post.CreatedAt,
				MarketImpact: market.Change24h,
				ImpactWindow: formatDuration(time.Since(post.CreatedAt)),
			}

			signals = append(signals, signal)
		}
	}

	// Limit results
	if len(signals) > c.config.MaxSignalsPerArticle {
		signals = signals[:c.config.MaxSignalsPerArticle]
	}

	return signals, nil
}

// FindRecentSignals finds all recent social signals that may have market impact.
func (c *Correlator) FindRecentSignals(ctx context.Context, lookback time.Duration) ([]models.SocialSignal, error) {
	users, err := c.GetTrackedUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting tracked users: %w", err)
	}

	var allSignals []models.SocialSignal
	since := time.Now().Add(-lookback)

	for _, user := range users {
		posts, err := c.client.GetRecentPosts(ctx, user.Handle, since, 100)
		if err != nil {
			log.Warn().Err(err).Str("handle", user.Handle).Msg("Failed to get posts")
			continue
		}

		for _, post := range posts {
			// Get markets that moved after this post
			movements, err := c.findMarketMovements(ctx, post, user)
			if err != nil {
				continue
			}

			if len(movements) == 0 {
				continue
			}

			// Calculate average impact
			totalImpact := 0.0
			for _, m := range movements {
				totalImpact += math.Abs(m.Change)
			}
			avgImpact := totalImpact / float64(len(movements))

			signal := models.SocialSignal{
				Handle:          user.Handle,
				Name:            user.Name,
				AvatarURL:       user.AvatarURL,
				Verified:        user.Verified,
				Content:         truncateContent(post.Content, 280),
				TweetURL:        post.TweetURL(user.Handle),
				PostedAt:        post.CreatedAt,
				MarketImpact:    avgImpact,
				ImpactWindow:    formatDuration(c.config.TimeWindow),
				AffectedMarkets: movements,
			}

			allSignals = append(allSignals, signal)
		}
	}

	// Sort by impact (highest first)
	sortByImpact(allSignals)

	return allSignals, nil
}

// EnrichArticleWithSignals adds relevant social signals to an article.
func (c *Correlator) EnrichArticleWithSignals(ctx context.Context, article *models.Article) error {
	if article.PrimaryMarket == nil && len(article.Markets) == 0 {
		return nil
	}

	// Determine main market for correlation
	var primaryMarketSlug string
	if article.PrimaryMarket != nil {
		primaryMarketSlug = article.PrimaryMarket.Slug
	} else if len(article.Markets) > 0 {
		primaryMarketSlug = article.Markets[0].Slug
	}

	// Get market data
	market, err := c.store.GetMarketBySlug(ctx, primaryMarketSlug)
	if err != nil {
		return fmt.Errorf("getting market: %w", err)
	}

	// Find relevant signals
	signals, err := c.FindSignalsForMarket(ctx, market, 4*time.Hour)
	if err != nil {
		return fmt.Errorf("finding signals: %w", err)
	}

	if len(signals) > 0 {
		// Add to article's social signals
		article.SocialSignals = signals

		// Add handles to enrichment sources
		for _, sig := range signals {
			source := fmt.Sprintf("@%s (%s)", sig.Handle, sig.PostedAt.Format("Jan 2"))
			article.EnrichmentSources = append(article.EnrichmentSources, source)
		}

		log.Info().
			Str("article", article.Slug).
			Int("signals", len(signals)).
			Msg("Enriched article with social signals")
	}

	return nil
}

// findMarketMovements finds markets that moved significantly after a post.
func (c *Correlator) findMarketMovements(ctx context.Context, post Post, user TrackedUser) ([]models.MarketMovement, error) {
	var movements []models.MarketMovement

	// Get markets in relevant categories
	for _, category := range c.config.Categories {
		markets, err := c.store.GetMarketsByCategory(ctx, category, 20)
		if err != nil {
			continue
		}

		for _, market := range markets {
			// Check if market moved significantly
			change := market.Change24h
			if math.Abs(change) < c.config.MinMarketChange {
				continue
			}

			// Check content relevance (basic keyword matching)
			if !c.isContentRelevantToMarket(post.Content, &market) {
				continue
			}

			movement := models.MarketMovement{
				MarketSlug:  market.Slug,
				MarketTitle: market.Question,
				Category:    market.Category,
				ProbBefore:  market.Probability - change,
				ProbAfter:   market.Probability,
				Change:      change,
				TimeDelta:   "within " + formatDuration(c.config.TimeWindow),
			}

			movements = append(movements, movement)
		}
	}

	return movements, nil
}

// isRelevant checks if a post is relevant to a market.
func (c *Correlator) isRelevant(content string, market *models.Market) bool {
	return c.isContentRelevantToMarket(content, market)
}

// isContentRelevantToMarket performs basic keyword matching.
func (c *Correlator) isContentRelevantToMarket(content string, market *models.Market) bool {
	contentLower := strings.ToLower(content)
	questionLower := strings.ToLower(market.Question)

	// Extract keywords from market question
	keywords := extractKeywords(questionLower)

	// Check if any keyword appears in the content
	matchCount := 0
	for _, kw := range keywords {
		if len(kw) > 3 && strings.Contains(contentLower, kw) {
			matchCount++
		}
	}

	// Require at least 2 keyword matches for relevance
	return matchCount >= 2
}

// isTimeCorrelated checks if the post timing correlates with market movement.
func (c *Correlator) isTimeCorrelated(postTime time.Time, market *models.Market) bool {
	// For now, just check if post is within our lookback window
	// Future: check market price history for movement after post time
	return time.Since(postTime) <= c.config.TimeWindow
}

// Helper functions

func extractKeywords(text string) []string {
	// Remove common words and punctuation
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"is": true, "are": true, "was": true, "were": true, "be": true, "been": true,
		"will": true, "would": true, "could": true, "should": true, "may": true, "might": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true, "with": true,
		"by": true, "from": true, "as": true, "into": true, "through": true,
		"this": true, "that": true, "these": true, "those": true,
		"it": true, "its": true, "their": true, "they": true, "them": true,
		"what": true, "when": true, "where": true, "who": true, "which": true, "how": true,
		"if": true, "then": true, "else": true, "than": true,
	}

	// Clean and split
	text = strings.ReplaceAll(text, "?", " ")
	text = strings.ReplaceAll(text, "'", " ")
	text = strings.ReplaceAll(text, "\"", " ")
	words := strings.Fields(text)

	var keywords []string
	for _, w := range words {
		w = strings.Trim(w, ".,!?:;\"'()[]")
		if len(w) > 2 && !stopWords[w] {
			keywords = append(keywords, w)
		}
	}

	return keywords
}

func truncateContent(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func sortByImpact(signals []models.SocialSignal) {
	// Simple bubble sort for now
	for i := 0; i < len(signals)-1; i++ {
		for j := i + 1; j < len(signals); j++ {
			if signals[j].MarketImpact > signals[i].MarketImpact {
				signals[i], signals[j] = signals[j], signals[i]
			}
		}
	}
}

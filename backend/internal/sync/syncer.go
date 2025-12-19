// Package sync provides continuous market data synchronization.
package sync

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/leeaandrob/futuresignals/internal/models"
	"github.com/leeaandrob/futuresignals/internal/polymarket"
	"github.com/leeaandrob/futuresignals/internal/storage"
	"github.com/rs/zerolog/log"
)

// Event types for the event bus.
type EventType string

const (
	EventNewMarket      EventType = "new_market"
	EventPriceChange    EventType = "price_change"
	EventBreakingMove   EventType = "breaking_move"
	EventVolumeSpike    EventType = "volume_spike"
	EventThresholdCross EventType = "threshold_cross"
	EventTrendingUpdate EventType = "trending_update"
)

// Event represents a market event.
type Event struct {
	Type      EventType
	Market    *models.Market
	Previous  *models.Snapshot
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// SyncerConfig holds configuration for the syncer.
type SyncerConfig struct {
	// How often to sync market data
	SyncInterval time.Duration

	// How often to take snapshots
	SnapshotInterval time.Duration

	// Thresholds for event detection
	BreakingThreshold   float64 // e.g., 0.05 = 5% change
	VolumeMultiplier    float64 // e.g., 3.0 = 3x normal volume
	TrendingThreshold   float64 // Minimum trending score

	// Cleanup
	SnapshotRetention time.Duration // How long to keep snapshots

	// Market filters
	MinVolume24h float64
}

// DefaultSyncerConfig returns default configuration.
func DefaultSyncerConfig() SyncerConfig {
	return SyncerConfig{
		SyncInterval:        30 * time.Second,
		SnapshotInterval:    5 * time.Minute,
		BreakingThreshold:   0.05,
		VolumeMultiplier:    3.0,
		TrendingThreshold:   50.0,
		SnapshotRetention:   7 * 24 * time.Hour,
		MinVolume24h:        10000,
	}
}

// Syncer continuously syncs market data from Polymarket.
type Syncer struct {
	client *polymarket.Client
	store  *storage.Store
	config SyncerConfig

	// Event channels
	events     chan Event
	eventMux   sync.RWMutex
	subscribers []chan Event

	// Market state cache
	marketCache   map[string]*models.Market
	cacheMux      sync.RWMutex

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewSyncer creates a new market syncer.
func NewSyncer(client *polymarket.Client, store *storage.Store, config SyncerConfig) *Syncer {
	ctx, cancel := context.WithCancel(context.Background())

	return &Syncer{
		client:      client,
		store:       store,
		config:      config,
		events:      make(chan Event, 1000),
		subscribers: make([]chan Event, 0),
		marketCache: make(map[string]*models.Market),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Subscribe returns a channel that receives market events.
func (s *Syncer) Subscribe() <-chan Event {
	s.eventMux.Lock()
	defer s.eventMux.Unlock()

	ch := make(chan Event, 100)
	s.subscribers = append(s.subscribers, ch)
	return ch
}

// Start begins the sync loops.
func (s *Syncer) Start() {
	log.Info().
		Dur("sync_interval", s.config.SyncInterval).
		Dur("snapshot_interval", s.config.SnapshotInterval).
		Msg("Starting market syncer")

	// Load existing markets into cache
	s.loadMarketCache()

	// Start the main sync loop
	s.wg.Add(1)
	go s.syncLoop()

	// Start the snapshot loop
	s.wg.Add(1)
	go s.snapshotLoop()

	// Start the event dispatcher
	s.wg.Add(1)
	go s.eventDispatcher()

	// Start the cleanup loop
	s.wg.Add(1)
	go s.cleanupLoop()
}

// Stop stops the syncer.
func (s *Syncer) Stop() {
	log.Info().Msg("Stopping market syncer")
	s.cancel()
	s.wg.Wait()
	close(s.events)

	// Close subscriber channels
	s.eventMux.Lock()
	for _, ch := range s.subscribers {
		close(ch)
	}
	s.eventMux.Unlock()
}

// loadMarketCache loads existing markets into the cache.
func (s *Syncer) loadMarketCache() {
	markets, err := s.store.GetAllActiveMarkets(s.ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load market cache")
		return
	}

	s.cacheMux.Lock()
	defer s.cacheMux.Unlock()

	for i := range markets {
		s.marketCache[markets[i].MarketID] = &markets[i]
	}

	log.Info().Int("markets", len(markets)).Msg("Loaded market cache")
}

// syncLoop continuously syncs market data.
func (s *Syncer) syncLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.SyncInterval)
	defer ticker.Stop()

	// Initial sync
	s.syncMarkets()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.syncMarkets()
		}
	}
}

// syncMarkets fetches and processes market data.
func (s *Syncer) syncMarkets() {
	log.Debug().Msg("Syncing markets")

	// Fetch top events by volume to get correct event slugs for URLs
	active := true
	closed := false
	events, err := s.client.GetEvents(s.ctx, polymarket.EventFilters{
		Active:    &active,
		Closed:    &closed,
		Limit:     100,
		Order:     "volume24hr",
		Ascending: false,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch events")
		return
	}

	log.Debug().Int("count", len(events)).Msg("Fetched events from Polymarket")

	// Process all markets from events with correct event slugs and event volume
	for _, event := range events {
		for _, pm := range event.Markets {
			s.processMarketWithEvent(pm, event)
		}
	}

	// Update trending scores
	s.updateTrendingScores()
}

// processMarketWithEvent processes a single market update with full event data.
func (s *Syncer) processMarketWithEvent(pm polymarket.Market, event polymarket.Event) {
	// Skip low volume markets
	if pm.Volume24hr < s.config.MinVolume24h {
		return
	}

	// Convert to our model with event data (slug + volumes)
	market := s.convertMarketWithEvent(pm, event)

	// Check cache for existing market
	s.cacheMux.RLock()
	existing, exists := s.marketCache[market.MarketID]
	s.cacheMux.RUnlock()

	if !exists {
		// New market detected
		market.FirstSeenAt = time.Now()
		s.emitEvent(Event{
			Type:      EventNewMarket,
			Market:    market,
			Timestamp: time.Now(),
		})
	} else {
		// Calculate changes
		market.FirstSeenAt = existing.FirstSeenAt
		market.PreviousProb = existing.Probability
		market.Change24h = market.Probability - existing.Probability

		// Check for breaking move
		if abs(market.Change24h) >= s.config.BreakingThreshold {
			s.emitEvent(Event{
				Type:      EventBreakingMove,
				Market:    market,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"change":       market.Change24h,
					"previous":     existing.Probability,
					"current":      market.Probability,
				},
			})
		}

		// Check for volume spike
		if existing.Volume24h > 0 && market.Volume24h/existing.Volume24h >= s.config.VolumeMultiplier {
			s.emitEvent(Event{
				Type:      EventVolumeSpike,
				Market:    market,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"previous_volume": existing.Volume24h,
					"current_volume":  market.Volume24h,
					"multiplier":      market.Volume24h / existing.Volume24h,
				},
			})
		}

		// Check for threshold crossings (50%, 75%, 90%)
		thresholds := []float64{0.50, 0.75, 0.90}
		for _, t := range thresholds {
			if crossedThreshold(existing.Probability, market.Probability, t) {
				s.emitEvent(Event{
					Type:      EventThresholdCross,
					Market:    market,
					Timestamp: time.Now(),
					Metadata: map[string]interface{}{
						"threshold": t,
						"direction": directionString(existing.Probability, market.Probability),
					},
				})
			}
		}
	}

	// Update cache
	s.cacheMux.Lock()
	s.marketCache[market.MarketID] = market
	s.cacheMux.Unlock()

	// Save to database
	if err := s.store.UpsertMarket(s.ctx, market); err != nil {
		log.Error().Err(err).Str("market_id", market.MarketID).Msg("Failed to save market")
	}
}

// processMarket processes a single market update (legacy, without event slug).
func (s *Syncer) processMarket(pm polymarket.Market) {
	// Skip low volume markets
	if pm.Volume24hr < s.config.MinVolume24h {
		return
	}

	// Convert to our model (uses market slug as fallback)
	market := s.convertMarket(pm)

	// Check cache for existing market
	s.cacheMux.RLock()
	existing, exists := s.marketCache[market.MarketID]
	s.cacheMux.RUnlock()

	if !exists {
		// New market detected
		market.FirstSeenAt = time.Now()
		s.emitEvent(Event{
			Type:      EventNewMarket,
			Market:    market,
			Timestamp: time.Now(),
		})
	} else {
		// Calculate changes
		market.FirstSeenAt = existing.FirstSeenAt
		market.PreviousProb = existing.Probability
		market.Change24h = market.Probability - existing.Probability

		// Check for breaking move
		if abs(market.Change24h) >= s.config.BreakingThreshold {
			s.emitEvent(Event{
				Type:      EventBreakingMove,
				Market:    market,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"change":       market.Change24h,
					"previous":     existing.Probability,
					"current":      market.Probability,
				},
			})
		}

		// Check for volume spike
		if existing.Volume24h > 0 && market.Volume24h/existing.Volume24h >= s.config.VolumeMultiplier {
			s.emitEvent(Event{
				Type:      EventVolumeSpike,
				Market:    market,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"previous_volume": existing.Volume24h,
					"current_volume":  market.Volume24h,
					"multiplier":      market.Volume24h / existing.Volume24h,
				},
			})
		}

		// Check for threshold crossings (50%, 75%, 90%)
		thresholds := []float64{0.50, 0.75, 0.90}
		for _, t := range thresholds {
			if crossedThreshold(existing.Probability, market.Probability, t) {
				s.emitEvent(Event{
					Type:      EventThresholdCross,
					Market:    market,
					Timestamp: time.Now(),
					Metadata: map[string]interface{}{
						"threshold": t,
						"direction": directionString(existing.Probability, market.Probability),
					},
				})
			}
		}
	}

	// Update cache
	s.cacheMux.Lock()
	s.marketCache[market.MarketID] = market
	s.cacheMux.Unlock()

	// Save to database
	if err := s.store.UpsertMarket(s.ctx, market); err != nil {
		log.Error().Err(err).Str("market_id", market.MarketID).Msg("Failed to save market")
	}
}

// convertMarketWithEvent converts a Polymarket market to our model with full event data.
func (s *Syncer) convertMarketWithEvent(pm polymarket.Market, event polymarket.Event) *models.Market {
	// Convert outcome prices from strings to floats
	var outcomePrices []float64
	for _, p := range pm.OutcomePrices {
		if f, err := parseFloat(p); err == nil {
			outcomePrices = append(outcomePrices, f)
		}
	}

	// Convert Polymarket tags to our model
	var polymarketTags []models.PolymarketTag
	for _, tag := range event.Tags {
		polymarketTags = append(polymarketTags, models.PolymarketTag{
			Label: tag.Label,
			Slug:  tag.Slug,
		})
	}

	// Use market image if available, otherwise use event image
	image := pm.Image
	if image == "" {
		image = event.Image
	}
	icon := pm.Icon
	if icon == "" {
		icon = event.Icon
	}

	market := &models.Market{
		// Identifiers
		MarketID:       pm.ID,
		ConditionID:    pm.ConditionID,
		GroupItemTitle: pm.GroupItemTitle,

		// Content
		Question:    pm.Question,
		Description: pm.Description,
		Image:       image,
		Icon:        icon,

		// Pricing
		Probability:    pm.YesPrice,
		LastTradePrice: pm.LastTradePrice,
		Change24h:      pm.OneDayPriceChange,
		Change7d:       pm.OneWeekPriceChange,

		// Volume
		Volume24h:      pm.Volume24hr,
		Volume7d:       pm.Volume1wk,
		TotalVolume:    pm.VolumeNum,
		EventVolume:    event.Volume,
		EventVolume24h: event.Volume24hr,

		// Event data
		EventTitle:   event.Title,
		CommentCount: event.CommentCount,
		SeriesSlug:   event.SeriesSlug,

		// Classification
		PolymarketTags: polymarketTags,

		// Liquidity & Status
		Liquidity:    pm.LiquidityNum,
		Active:       pm.Active,
		Closed:       pm.Closed,
		Archived:     false,
		AcceptingBid: pm.AcceptingOrders,
		StartDate:    pm.StartDate,
		EndDate:      pm.EndDate,

		// Resolution
		ResolutionSource: pm.ResolutionSource,
		CompetitorCount:  event.CompetitorCount,

		// Outcomes
		Outcomes:      []string(pm.Outcomes),
		OutcomePrices: outcomePrices,

		// Meta
		UpdatedAt:     time.Now(),
		PolymarketURL: "https://polymarket.com/event/" + event.Slug,
	}

	// Detect category
	market.Category = market.DetectCategory()

	// Generate slug
	market.Slug = market.GenerateSlug()

	// Calculate trending score
	market.TrendingScore = market.CalculateTrendingScore()

	return market
}

// convertMarket converts a Polymarket market to our model (legacy, uses market slug as fallback).
func (s *Syncer) convertMarket(pm polymarket.Market) *models.Market {
	// Convert outcome prices from strings to floats
	var outcomePrices []float64
	for _, p := range pm.OutcomePrices {
		if f, err := parseFloat(p); err == nil {
			outcomePrices = append(outcomePrices, f)
		}
	}

	market := &models.Market{
		MarketID:       pm.ID,
		ConditionID:    pm.ConditionID,
		GroupItemTitle: pm.GroupItemTitle,
		Question:       pm.Question,
		Description:    pm.Description,
		Probability:    pm.YesPrice,
		Volume24h:      pm.Volume24hr,
		TotalVolume:    pm.VolumeNum,
		Liquidity:      pm.LiquidityNum,
		Active:         pm.Active,
		Closed:         pm.Closed,
		Archived:       false,
		AcceptingBid:   pm.AcceptingOrders,
		EndDate:        pm.EndDate,
		Outcomes:       []string(pm.Outcomes),
		OutcomePrices:  outcomePrices,
		UpdatedAt:      time.Now(),
		PolymarketURL:  "https://polymarket.com/event/" + pm.Slug,
	}

	// Detect category
	market.Category = market.DetectCategory()

	// Generate slug
	market.Slug = market.GenerateSlug()

	// Calculate trending score
	market.TrendingScore = market.CalculateTrendingScore()

	return market
}

// updateTrendingScores recalculates trending scores for all cached markets.
func (s *Syncer) updateTrendingScores() {
	s.cacheMux.Lock()
	defer s.cacheMux.Unlock()

	for _, market := range s.marketCache {
		market.TrendingScore = market.CalculateTrendingScore()
	}
}

// snapshotLoop takes periodic snapshots of market data.
func (s *Syncer) snapshotLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.SnapshotInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.takeSnapshots()
		}
	}
}

// takeSnapshots saves snapshots of all cached markets.
func (s *Syncer) takeSnapshots() {
	log.Debug().Msg("Taking market snapshots")

	s.cacheMux.RLock()
	defer s.cacheMux.RUnlock()

	for _, market := range s.marketCache {
		snapshot := &models.Snapshot{
			MarketID:    market.MarketID,
			Probability: market.Probability,
			Volume24h:   market.Volume24h,
			TotalVolume: market.TotalVolume,
			Liquidity:   market.Liquidity,
		}

		if err := s.store.SaveSnapshot(s.ctx, snapshot); err != nil {
			log.Error().Err(err).Str("market_id", market.MarketID).Msg("Failed to save snapshot")
		}
	}

	log.Debug().Int("count", len(s.marketCache)).Msg("Snapshots saved")
}

// cleanupLoop periodically cleans old data.
func (s *Syncer) cleanupLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

// cleanup removes old snapshots.
func (s *Syncer) cleanup() {
	deleted, err := s.store.CleanOldSnapshots(s.ctx, s.config.SnapshotRetention)
	if err != nil {
		log.Error().Err(err).Msg("Failed to clean old snapshots")
		return
	}

	if deleted > 0 {
		log.Info().Int64("deleted", deleted).Msg("Cleaned old snapshots")
	}
}

// eventDispatcher dispatches events to subscribers.
func (s *Syncer) eventDispatcher() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case event, ok := <-s.events:
			if !ok {
				return
			}

			s.eventMux.RLock()
			for _, sub := range s.subscribers {
				select {
				case sub <- event:
				default:
					log.Warn().Msg("Subscriber channel full, dropping event")
				}
			}
			s.eventMux.RUnlock()
		}
	}
}

// emitEvent sends an event to the event channel.
func (s *Syncer) emitEvent(event Event) {
	select {
	case s.events <- event:
		log.Debug().
			Str("type", string(event.Type)).
			Str("market", event.Market.Question).
			Msg("Event emitted")
	default:
		log.Warn().Msg("Event channel full, dropping event")
	}
}

// Helper functions

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func crossedThreshold(prev, curr, threshold float64) bool {
	return (prev < threshold && curr >= threshold) || (prev >= threshold && curr < threshold)
}

func directionString(prev, curr float64) string {
	if curr > prev {
		return "up"
	}
	return "down"
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// GetCachedMarket returns a market from the cache.
func (s *Syncer) GetCachedMarket(marketID string) (*models.Market, bool) {
	s.cacheMux.RLock()
	defer s.cacheMux.RUnlock()
	m, ok := s.marketCache[marketID]
	return m, ok
}

// SyncNow forces an immediate sync of market data.
func (s *Syncer) SyncNow() {
	log.Info().Msg("Manual sync triggered")
	s.syncMarkets()
}

// GetTrendingMarkets returns the top trending markets from cache.
func (s *Syncer) GetTrendingMarkets(limit int) []*models.Market {
	s.cacheMux.RLock()
	defer s.cacheMux.RUnlock()

	// Collect all markets
	markets := make([]*models.Market, 0, len(s.marketCache))
	for _, m := range s.marketCache {
		if m.Active && !m.Closed {
			markets = append(markets, m)
		}
	}

	// Sort by trending score (simple bubble sort for small sets)
	for i := 0; i < len(markets)-1; i++ {
		for j := i + 1; j < len(markets); j++ {
			if markets[j].TrendingScore > markets[i].TrendingScore {
				markets[i], markets[j] = markets[j], markets[i]
			}
		}
	}

	// Limit results
	if len(markets) > limit {
		markets = markets[:limit]
	}

	return markets
}

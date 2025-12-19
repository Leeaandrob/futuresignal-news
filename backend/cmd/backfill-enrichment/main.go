// Package main provides a backfill script to enrich markets with additional Polymarket data.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Tag from Polymarket API
type Tag struct {
	Label string `json:"label"`
	Slug  string `json:"slug"`
}

// Event from Polymarket API
type Event struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Slug             string   `json:"slug"`
	Image            string   `json:"image"`
	Icon             string   `json:"icon"`
	Volume           float64  `json:"volume"`
	Volume24hr       float64  `json:"volume24hr"`
	Volume1wk        float64  `json:"volume1wk"`
	CommentCount     int      `json:"commentCount"`
	CompetitorCount  int      `json:"competitorCount"`
	SeriesSlug       string   `json:"seriesSlug"`
	ResolutionSource string   `json:"resolutionSource"`
	Tags             []Tag    `json:"tags"`
	Markets          []Market `json:"markets"`
}

// Market from Polymarket API
type Market struct {
	ID                 string   `json:"id"`
	Image              string   `json:"image"`
	Icon               string   `json:"icon"`
	StartDate          string   `json:"startDate"`
	OutcomePrices      []string `json:"outcomePrices"`
	VolumeNum          float64  `json:"volumeNum"`
	Volume24hr         float64  `json:"volume24hr"`
	Volume1wk          float64  `json:"volume1wk"`
	LiquidityNum       float64  `json:"liquidityNum"`
	LastTradePrice     float64  `json:"lastTradePrice"`
	OneDayPriceChange  float64  `json:"oneDayPriceChange"`
	OneWeekPriceChange float64  `json:"oneWeekPriceChange"`
	ResolutionSource   string   `json:"resolutionSource"`
}

// UnmarshalJSON handles the JSON string array for OutcomePrices
func (m *Market) UnmarshalJSON(data []byte) error {
	type Alias Market
	aux := &struct {
		OutcomePrices interface{} `json:"outcomePrices"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.OutcomePrices.(type) {
	case string:
		if v != "" {
			var prices []string
			if err := json.Unmarshal([]byte(v), &prices); err == nil {
				m.OutcomePrices = prices
			}
		}
	case []interface{}:
		for _, p := range v {
			if s, ok := p.(string); ok {
				m.OutcomePrices = append(m.OutcomePrices, s)
			}
		}
	}

	return nil
}

// PolymarketTag for MongoDB
type PolymarketTag struct {
	Label string `bson:"label"`
	Slug  string `bson:"slug"`
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal().Msg("MONGODB_URI environment variable is required")
	}

	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "futuresignals"
	}

	log.Info().Msg("Starting market enrichment backfill")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer client.Disconnect(ctx)

	collection := client.Database(dbName).Collection("markets")

	// Get all markets from our database
	log.Info().Msg("Fetching markets from database...")

	type DBMarket struct {
		MarketID string `bson:"market_id"`
	}

	var markets []DBMarket
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to query markets")
	}

	for cursor.Next(ctx) {
		var m DBMarket
		if err := cursor.Decode(&m); err == nil && m.MarketID != "" {
			markets = append(markets, m)
		}
	}
	cursor.Close(ctx)

	log.Info().Int("count", len(markets)).Msg("Found markets in database")

	if len(markets) == 0 {
		log.Info().Msg("No markets in database, nothing to update")
		return
	}

	// Fetch events from Polymarket
	httpClient := &http.Client{Timeout: 30 * time.Second}

	log.Info().Msg("Fetching events from Polymarket API...")

	type EnrichmentData struct {
		YesPrice           float64
		LastTradePrice     float64
		OneDayPriceChange  float64
		OneWeekPriceChange float64
		TotalVolume        float64
		Volume24h          float64
		Volume7d           float64
		Liquidity          float64
		EventVolume        float64
		EventVolume24h     float64
		EventTitle         string
		Image              string
		Icon               string
		CommentCount       int
		CompetitorCount    int
		SeriesSlug         string
		ResolutionSource   string
		StartDate          string
		Tags               []PolymarketTag
	}
	marketDataMap := make(map[string]EnrichmentData)

	// Fetch top 100 events by volume
	resp, err := httpClient.Get("https://gamma-api.polymarket.com/events?active=true&closed=false&limit=100&order=volume24hr&ascending=false")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to fetch events")
	}
	defer resp.Body.Close()

	var events []Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		log.Fatal().Err(err).Msg("Failed to decode events")
	}

	log.Info().Int("count", len(events)).Msg("Fetched events from Polymarket")

	// Build market map with enrichment data
	for _, event := range events {
		// Convert event tags
		var tags []PolymarketTag
		for _, t := range event.Tags {
			tags = append(tags, PolymarketTag{Label: t.Label, Slug: t.Slug})
		}

		for _, market := range event.Markets {
			yesPrice := 0.0
			if len(market.OutcomePrices) >= 1 {
				yesPrice, _ = strconv.ParseFloat(market.OutcomePrices[0], 64)
			}

			// Use market image if available, otherwise event image
			image := market.Image
			if image == "" {
				image = event.Image
			}
			icon := market.Icon
			if icon == "" {
				icon = event.Icon
			}

			marketDataMap[market.ID] = EnrichmentData{
				YesPrice:           yesPrice,
				LastTradePrice:     market.LastTradePrice,
				OneDayPriceChange:  market.OneDayPriceChange,
				OneWeekPriceChange: market.OneWeekPriceChange,
				TotalVolume:        market.VolumeNum,
				Volume24h:          market.Volume24hr,
				Volume7d:           market.Volume1wk,
				Liquidity:          market.LiquidityNum,
				EventVolume:        event.Volume,
				EventVolume24h:     event.Volume24hr,
				EventTitle:         event.Title,
				Image:              image,
				Icon:               icon,
				CommentCount:       event.CommentCount,
				CompetitorCount:    event.CompetitorCount,
				SeriesSlug:         event.SeriesSlug,
				ResolutionSource:   market.ResolutionSource,
				StartDate:          market.StartDate,
				Tags:               tags,
			}
		}
	}

	log.Info().Int("count", len(marketDataMap)).Msg("Built enrichment data map")

	// Update each market
	updated := 0
	skipped := 0
	errors := 0

	for i, m := range markets {
		data, found := marketDataMap[m.MarketID]
		if !found {
			skipped++
			continue
		}

		update := bson.M{
			"$set": bson.M{
				// Pricing
				"probability":       data.YesPrice,
				"last_trade_price":  data.LastTradePrice,
				"change_24h":        data.OneDayPriceChange,
				"change_7d":         data.OneWeekPriceChange,

				// Volume
				"volume_24h":        data.Volume24h,
				"volume_7d":         data.Volume7d,
				"total_volume":      data.TotalVolume,
				"event_volume":      data.EventVolume,
				"event_volume_24h":  data.EventVolume24h,

				// Event data
				"event_title":       data.EventTitle,
				"comment_count":     data.CommentCount,
				"series_slug":       data.SeriesSlug,

				// Media
				"image":             data.Image,
				"icon":              data.Icon,

				// Resolution
				"resolution_source": data.ResolutionSource,
				"competitor_count":  data.CompetitorCount,

				// Classification
				"polymarket_tags":   data.Tags,

				// Status
				"start_date":        data.StartDate,
				"liquidity":         data.Liquidity,

				// Meta
				"updated_at":        time.Now(),
			},
		}

		_, err = collection.UpdateOne(
			ctx,
			bson.M{"market_id": m.MarketID},
			update,
		)
		if err != nil {
			log.Error().Err(err).Str("market_id", m.MarketID).Msg("Failed to update market")
			errors++
			continue
		}

		updated++

		if (i+1)%50 == 0 {
			log.Info().
				Int("processed", i+1).
				Int("total", len(markets)).
				Int("updated", updated).
				Int("skipped", skipped).
				Msg("Progress")
		}

		time.Sleep(10 * time.Millisecond)
	}

	fmt.Printf("\n✅ Enrichment backfill complete!\n")
	fmt.Printf("   Updated: %d markets\n", updated)
	fmt.Printf("   Skipped: %d markets (not in top 100 events)\n", skipped)
	fmt.Printf("   Errors: %d\n", errors)
}

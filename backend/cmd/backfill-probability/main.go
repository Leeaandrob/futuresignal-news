// Package main provides a backfill script to fix market probability and add event volume.
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

// Event from Polymarket API
type Event struct {
	ID         string    `json:"id"`
	Slug       string    `json:"slug"`
	Volume     float64   `json:"volume"`
	Volume24hr float64   `json:"volume24hr"`
	Markets    []Market  `json:"markets"`
}

// Market from Polymarket API
type Market struct {
	ID            string   `json:"id"`
	OutcomePrices []string `json:"outcomePrices"`
	VolumeNum     float64  `json:"volumeNum"`
	Volume24hr    float64  `json:"volume24hr"`
	LiquidityNum  float64  `json:"liquidityNum"`
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

	// Handle outcomePrices as string or array
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

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Get MongoDB URI from environment
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal().Msg("MONGODB_URI environment variable is required")
	}

	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "futuresignals"
	}

	log.Info().Msg("Starting probability and event volume backfill")

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer client.Disconnect(ctx)

	collection := client.Database(dbName).Collection("markets")

	// Step 1: Get all markets from our database
	log.Info().Msg("Fetching markets from database...")

	type DBMarket struct {
		MarketID    string  `bson:"market_id"`
		Probability float64 `bson:"probability"`
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

	// Step 2: Fetch events from Polymarket to get correct probabilities
	httpClient := &http.Client{Timeout: 30 * time.Second}
	updated := 0
	skipped := 0
	errors := 0

	// Build a map of market_id -> event data
	log.Info().Msg("Fetching events from Polymarket API...")

	type EventData struct {
		YesPrice       float64
		EventVolume    float64
		EventVolume24h float64
		TotalVolume    float64
		Liquidity      float64
	}
	marketEventMap := make(map[string]EventData)

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

	// Build market map
	for _, event := range events {
		for _, market := range event.Markets {
			yesPrice := 0.0
			if len(market.OutcomePrices) >= 1 {
				yesPrice, _ = strconv.ParseFloat(market.OutcomePrices[0], 64)
			}
			marketEventMap[market.ID] = EventData{
				YesPrice:       yesPrice,
				EventVolume:    event.Volume,
				EventVolume24h: event.Volume24hr,
				TotalVolume:    market.VolumeNum,
				Liquidity:      market.LiquidityNum,
			}
		}
	}

	log.Info().Int("count", len(marketEventMap)).Msg("Built market event map")

	// Step 3: Update each market
	for i, m := range markets {
		eventData, found := marketEventMap[m.MarketID]
		if !found {
			skipped++
			continue
		}

		// Update with correct data
		update := bson.M{
			"$set": bson.M{
				"probability":       eventData.YesPrice,
				"total_volume":      eventData.TotalVolume,
				"liquidity":         eventData.Liquidity,
				"event_volume":      eventData.EventVolume,
				"event_volume_24h":  eventData.EventVolume24h,
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
		if eventData.YesPrice != m.Probability {
			log.Debug().
				Str("market_id", m.MarketID).
				Float64("old_prob", m.Probability).
				Float64("new_prob", eventData.YesPrice).
				Float64("event_volume", eventData.EventVolume).
				Msg("Updated market")
		}

		// Progress log every 50 markets
		if (i+1)%50 == 0 {
			log.Info().
				Int("processed", i+1).
				Int("total", len(markets)).
				Int("updated", updated).
				Int("skipped", skipped).
				Msg("Progress")
		}

		// Rate limiting
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Printf("\n✅ Backfill complete!\n")
	fmt.Printf("   Updated: %d markets\n", updated)
	fmt.Printf("   Skipped (not in Polymarket response): %d markets\n", skipped)
	fmt.Printf("   Errors: %d\n", errors)
}

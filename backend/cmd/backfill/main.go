// Package main provides a backfill script to fix Polymarket URLs.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MarketWithEvents is a simplified response from Polymarket markets API
type MarketWithEvents struct {
	ID     string `json:"id"`
	Events []struct {
		Slug string `json:"slug"`
	} `json:"events"`
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

	log.Info().Msg("Starting Polymarket URL backfill")

	// Connect to MongoDB with longer timeout for backfill operations
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
		MarketID      string `bson:"market_id"`
		PolymarketURL string `bson:"polymarket_url"`
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

	// Step 2: For each market, fetch from Polymarket API to get event slug
	httpClient := &http.Client{Timeout: 10 * time.Second}
	updated := 0
	skipped := 0
	notFound := 0
	errors := 0

	for i, m := range markets {
		// Fetch market from Polymarket by ID
		url := fmt.Sprintf("https://gamma-api.polymarket.com/markets?id=%s", m.MarketID)
		resp, err := httpClient.Get(url)
		if err != nil {
			log.Error().Err(err).Str("market_id", m.MarketID).Msg("Failed to fetch from Polymarket")
			errors++
			continue
		}

		var pmMarkets []MarketWithEvents
		if err := json.NewDecoder(resp.Body).Decode(&pmMarkets); err != nil {
			resp.Body.Close()
			log.Error().Err(err).Str("market_id", m.MarketID).Msg("Failed to decode response")
			errors++
			continue
		}
		resp.Body.Close()

		if len(pmMarkets) == 0 || len(pmMarkets[0].Events) == 0 {
			notFound++
			continue
		}

		eventSlug := pmMarkets[0].Events[0].Slug
		newURL := "https://polymarket.com/event/" + eventSlug

		// Only update if URL is different
		if m.PolymarketURL == newURL {
			skipped++
			if (i+1)%50 == 0 {
				log.Info().
					Int("processed", i+1).
					Int("total", len(markets)).
					Int("updated", updated).
					Int("skipped", skipped).
					Msg("Progress")
			}
			continue
		}

		_, err = collection.UpdateOne(
			ctx,
			bson.M{"market_id": m.MarketID},
			bson.M{"$set": bson.M{"polymarket_url": newURL}},
		)
		if err != nil {
			log.Error().Err(err).Str("market_id", m.MarketID).Msg("Failed to update market")
			errors++
			continue
		}

		updated++
		log.Debug().
			Str("market_id", m.MarketID).
			Str("new_url", newURL).
			Msg("Updated market URL")

		// Progress log every 50 markets
		if (i+1)%50 == 0 {
			log.Info().
				Int("processed", i+1).
				Int("total", len(markets)).
				Int("updated", updated).
				Int("skipped", skipped).
				Msg("Progress")
		}

		// Rate limiting: 10 requests per second
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\n✅ Backfill complete!\n")
	fmt.Printf("   Updated: %d markets\n", updated)
	fmt.Printf("   Skipped (already correct): %d markets\n", skipped)
	fmt.Printf("   Not found in Polymarket: %d markets\n", notFound)
	fmt.Printf("   Errors: %d\n", errors)
}

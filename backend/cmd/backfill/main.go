// Package main provides a backfill script to fix Polymarket URLs.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/leeaandrob/futuresignals/internal/polymarket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer client.Disconnect(ctx)

	collection := client.Database(dbName).Collection("markets")

	// Create Polymarket client
	pmClient := polymarket.NewClient()

	// Fetch all events from Polymarket
	log.Info().Msg("Fetching events from Polymarket...")

	// Build market_id -> event_slug mapping
	marketToEvent := make(map[string]string)
	offset := 0
	limit := 100

	for {
		events, err := pmClient.GetEvents(ctx, polymarket.EventFilters{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to fetch events")
			break
		}

		if len(events) == 0 {
			break
		}

		for _, event := range events {
			for _, market := range event.Markets {
				marketToEvent[market.ID] = event.Slug
			}
		}

		log.Info().
			Int("offset", offset).
			Int("events", len(events)).
			Int("mappings", len(marketToEvent)).
			Msg("Fetched events batch")

		offset += limit

		// Rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	log.Info().Int("total_mappings", len(marketToEvent)).Msg("Finished fetching events")

	// Update markets in MongoDB
	updated := 0
	skipped := 0

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to query markets")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var market struct {
			MarketID      string `bson:"market_id"`
			PolymarketURL string `bson:"polymarket_url"`
		}

		if err := cursor.Decode(&market); err != nil {
			log.Error().Err(err).Msg("Failed to decode market")
			continue
		}

		eventSlug, ok := marketToEvent[market.MarketID]
		if !ok {
			skipped++
			continue
		}

		newURL := "https://polymarket.com/event/" + eventSlug

		// Only update if URL is different
		if market.PolymarketURL == newURL {
			skipped++
			continue
		}

		_, err := collection.UpdateOne(
			ctx,
			bson.M{"market_id": market.MarketID},
			bson.M{"$set": bson.M{"polymarket_url": newURL}},
		)
		if err != nil {
			log.Error().Err(err).Str("market_id", market.MarketID).Msg("Failed to update market")
			continue
		}

		updated++
		log.Debug().
			Str("market_id", market.MarketID).
			Str("new_url", newURL).
			Msg("Updated market URL")
	}

	fmt.Printf("\n✅ Backfill complete!\n")
	fmt.Printf("   Updated: %d markets\n", updated)
	fmt.Printf("   Skipped: %d markets\n", skipped)
}

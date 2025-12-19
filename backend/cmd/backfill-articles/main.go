// Package main provides a backfill script to update MarketRef data in existing articles.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MarketRef references a market within an article.
type MarketRef struct {
	MarketID     string  `bson:"market_id"`
	Question     string  `bson:"question"`
	Slug         string  `bson:"slug"`
	Probability  float64 `bson:"probability"`
	PreviousProb float64 `bson:"previous_prob,omitempty"`
	Change24h    float64 `bson:"change_24h"`
	Volume24h    float64 `bson:"volume_24h"`
	TotalVolume  float64 `bson:"total_volume"`
	EndDate      string  `bson:"end_date,omitempty"`
}

// Article for reading from DB.
type Article struct {
	ID            primitive.ObjectID `bson:"_id"`
	Headline      string             `bson:"headline"`
	Markets       []MarketRef        `bson:"markets"`
	PrimaryMarket *MarketRef         `bson:"primary_market,omitempty"`
}

// Market from database.
type Market struct {
	MarketID    string  `bson:"market_id"`
	Question    string  `bson:"question"`
	Slug        string  `bson:"slug"`
	Probability float64 `bson:"probability"`
	Change24h   float64 `bson:"change_24h"`
	Volume24h   float64 `bson:"volume_24h"`
	TotalVolume float64 `bson:"total_volume"`
	EndDate     string  `bson:"end_date"`
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

	log.Info().Msg("Starting articles MarketRef backfill")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer client.Disconnect(ctx)

	db := client.Database(dbName)
	articlesCol := db.Collection("articles")
	marketsCol := db.Collection("markets")

	// Get all articles
	log.Info().Msg("Fetching articles from database...")

	var articles []Article
	cursor, err := articlesCol.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to query articles")
	}

	if err := cursor.All(ctx, &articles); err != nil {
		log.Fatal().Err(err).Msg("Failed to decode articles")
	}

	log.Info().Int("count", len(articles)).Msg("Found articles in database")

	if len(articles) == 0 {
		log.Info().Msg("No articles in database, nothing to update")
		return
	}

	// Build a map of all market IDs we need
	marketIDs := make(map[string]bool)
	for _, article := range articles {
		for _, m := range article.Markets {
			marketIDs[m.MarketID] = true
		}
		if article.PrimaryMarket != nil {
			marketIDs[article.PrimaryMarket.MarketID] = true
		}
	}

	log.Info().Int("unique_markets", len(marketIDs)).Msg("Collecting market data")

	// Fetch all needed markets in one query
	var marketIDList []string
	for id := range marketIDs {
		marketIDList = append(marketIDList, id)
	}

	marketCursor, err := marketsCol.Find(ctx, bson.M{"market_id": bson.M{"$in": marketIDList}})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to query markets")
	}

	var markets []Market
	if err := marketCursor.All(ctx, &markets); err != nil {
		log.Fatal().Err(err).Msg("Failed to decode markets")
	}

	// Build lookup map
	marketMap := make(map[string]Market)
	for _, m := range markets {
		marketMap[m.MarketID] = m
	}

	log.Info().Int("found", len(marketMap)).Msg("Loaded market data")

	// Update each article
	updated := 0
	skipped := 0
	errors := 0

	for i, article := range articles {
		needsUpdate := false

		// Update Markets array
		var updatedMarkets []MarketRef
		for _, ref := range article.Markets {
			if m, found := marketMap[ref.MarketID]; found {
				updatedMarkets = append(updatedMarkets, MarketRef{
					MarketID:     m.MarketID,
					Question:     m.Question,
					Slug:         m.Slug,
					Probability:  m.Probability,
					PreviousProb: ref.PreviousProb, // Keep historical
					Change24h:    m.Change24h,
					Volume24h:    m.Volume24h,
					TotalVolume:  m.TotalVolume,
					EndDate:      m.EndDate,
				})
				needsUpdate = true
			} else {
				// Keep original if market not found
				updatedMarkets = append(updatedMarkets, ref)
			}
		}

		// Update PrimaryMarket
		var updatedPrimary *MarketRef
		if article.PrimaryMarket != nil {
			if m, found := marketMap[article.PrimaryMarket.MarketID]; found {
				updatedPrimary = &MarketRef{
					MarketID:     m.MarketID,
					Question:     m.Question,
					Slug:         m.Slug,
					Probability:  m.Probability,
					PreviousProb: article.PrimaryMarket.PreviousProb,
					Change24h:    m.Change24h,
					Volume24h:    m.Volume24h,
					TotalVolume:  m.TotalVolume,
					EndDate:      m.EndDate,
				}
				needsUpdate = true
			} else {
				updatedPrimary = article.PrimaryMarket
			}
		}

		if !needsUpdate {
			skipped++
			continue
		}

		// Update the article
		update := bson.M{
			"$set": bson.M{
				"markets":        updatedMarkets,
				"primary_market": updatedPrimary,
				"updated_at":     time.Now(),
			},
		}

		_, err = articlesCol.UpdateOne(ctx, bson.M{"_id": article.ID}, update)
		if err != nil {
			log.Error().Err(err).Str("article_id", article.ID.Hex()).Msg("Failed to update article")
			errors++
			continue
		}

		updated++
		log.Info().
			Int("progress", i+1).
			Int("total", len(articles)).
			Str("headline", truncate(article.Headline, 50)).
			Int("markets", len(updatedMarkets)).
			Msg("Updated article")
	}

	fmt.Printf("\n✅ Articles MarketRef backfill complete!\n")
	fmt.Printf("   Updated: %d articles\n", updated)
	fmt.Printf("   Skipped: %d articles (no markets found)\n", skipped)
	fmt.Printf("   Errors: %d\n", errors)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Package main fixes article slugs with problematic characters.
package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Article struct {
	ID   primitive.ObjectID `bson:"_id"`
	Slug string             `bson:"slug"`
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

	log.Info().Msg("Starting slug fix for articles")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer client.Disconnect(ctx)

	collection := client.Database(dbName).Collection("articles")

	// Find articles with problematic characters
	badCharsRegex := regexp.MustCompile(`[%$@#\+\[\]]`)

	cursor, err := collection.Find(ctx, bson.M{
		"slug": bson.M{"$regex": `[%$@#\+\[\]]`},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to query articles")
	}

	var articles []Article
	if err := cursor.All(ctx, &articles); err != nil {
		log.Fatal().Err(err).Msg("Failed to decode articles")
	}

	log.Info().Int("count", len(articles)).Msg("Found articles with bad slugs")

	if len(articles) == 0 {
		log.Info().Msg("No articles need fixing")
		return
	}

	// Fix each article
	fixed := 0
	for _, article := range articles {
		oldSlug := article.Slug
		newSlug := badCharsRegex.ReplaceAllString(oldSlug, "")
		// Clean up double dashes and trailing dashes
		newSlug = regexp.MustCompile(`-+`).ReplaceAllString(newSlug, "-")
		newSlug = strings.TrimRight(newSlug, "-")

		log.Info().
			Str("old", oldSlug).
			Str("new", newSlug).
			Msg("Fixing slug")

		_, err := collection.UpdateOne(ctx,
			bson.M{"_id": article.ID},
			bson.M{"$set": bson.M{
				"slug":       newSlug,
				"updated_at": time.Now(),
			}},
		)
		if err != nil {
			log.Error().Err(err).Str("slug", oldSlug).Msg("Failed to update")
			continue
		}
		fixed++
	}

	fmt.Printf("\n✅ Fixed %d article slugs\n", fixed)
}

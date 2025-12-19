// Package storage provides MongoDB storage for FutureSignals.
package storage

import (
	"context"
	"time"

	"github.com/leeaandrob/futuresignals/internal/models"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Store provides access to all MongoDB collections.
type Store struct {
	client     *mongo.Client
	db         *mongo.Database
	markets    *mongo.Collection
	snapshots  *mongo.Collection
	articles   *mongo.Collection
	categories *mongo.Collection
}

// NewStore creates a new storage connection.
func NewStore(ctx context.Context, uri, dbName string) (*Store, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(dbName)
	log.Info().Str("db", dbName).Msg("Connected to MongoDB")

	store := &Store{
		client:     client,
		db:         db,
		markets:    db.Collection("markets"),
		snapshots:  db.Collection("snapshots"),
		articles:   db.Collection("articles"),
		categories: db.Collection("categories"),
	}

	// Initialize indexes
	if err := store.createIndexes(ctx); err != nil {
		log.Warn().Err(err).Msg("Failed to create some indexes")
	}

	// Initialize default categories
	if err := store.initCategories(ctx); err != nil {
		log.Warn().Err(err).Msg("Failed to initialize categories")
	}

	return store, nil
}

// Close closes the database connection.
func (s *Store) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}

// createIndexes creates necessary indexes for efficient queries.
func (s *Store) createIndexes(ctx context.Context) error {
	// Markets indexes
	marketIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "market_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "category", Value: 1}}},
		{Keys: bson.D{{Key: "trending_score", Value: -1}}},
		{Keys: bson.D{{Key: "volume_24h", Value: -1}}},
		{Keys: bson.D{{Key: "change_24h", Value: -1}}},
		{Keys: bson.D{{Key: "first_seen_at", Value: -1}}},
		{Keys: bson.D{{Key: "active", Value: 1}}},
	}
	if _, err := s.markets.Indexes().CreateMany(ctx, marketIndexes); err != nil {
		log.Warn().Err(err).Msg("Failed to create market indexes")
	}

	// Snapshots indexes
	snapshotIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "market_id", Value: 1}, {Key: "captured_at", Value: -1}}},
		{Keys: bson.D{{Key: "captured_at", Value: -1}}},
	}
	if _, err := s.snapshots.Indexes().CreateMany(ctx, snapshotIndexes); err != nil {
		log.Warn().Err(err).Msg("Failed to create snapshot indexes")
	}

	// Articles indexes
	articleIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "type", Value: 1}}},
		{Keys: bson.D{{Key: "category", Value: 1}}},
		{Keys: bson.D{{Key: "published_at", Value: -1}}},
		{Keys: bson.D{{Key: "published", Value: 1}}},
		{Keys: bson.D{{Key: "featured", Value: 1}}},
		{Keys: bson.D{{Key: "tags", Value: 1}}},
	}
	if _, err := s.articles.Indexes().CreateMany(ctx, articleIndexes); err != nil {
		log.Warn().Err(err).Msg("Failed to create article indexes")
	}

	return nil
}

// initCategories initializes default categories if not present.
func (s *Store) initCategories(ctx context.Context) error {
	for _, cat := range models.DefaultCategories {
		filter := bson.M{"slug": cat.Slug}
		update := bson.M{"$setOnInsert": cat}
		opts := options.Update().SetUpsert(true)
		if _, err := s.categories.UpdateOne(ctx, filter, update, opts); err != nil {
			return err
		}
	}
	return nil
}

// ============================================================================
// MARKET OPERATIONS
// ============================================================================

// UpsertMarket inserts or updates a market.
func (s *Store) UpsertMarket(ctx context.Context, market *models.Market) error {
	market.UpdatedAt = time.Now()
	if market.FirstSeenAt.IsZero() {
		market.FirstSeenAt = time.Now()
	}

	filter := bson.M{"market_id": market.MarketID}
	update := bson.M{"$set": market}
	opts := options.Update().SetUpsert(true)

	_, err := s.markets.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetMarketByID returns a market by its Polymarket ID.
func (s *Store) GetMarketByID(ctx context.Context, marketID string) (*models.Market, error) {
	var market models.Market
	err := s.markets.FindOne(ctx, bson.M{"market_id": marketID}).Decode(&market)
	if err != nil {
		return nil, err
	}
	return &market, nil
}

// GetMarketBySlug returns a market by its slug.
func (s *Store) GetMarketBySlug(ctx context.Context, slug string) (*models.Market, error) {
	var market models.Market
	err := s.markets.FindOne(ctx, bson.M{"slug": slug}).Decode(&market)
	if err != nil {
		return nil, err
	}
	return &market, nil
}

// GetTrendingMarkets returns markets sorted by trending score.
func (s *Store) GetTrendingMarkets(ctx context.Context, limit int) ([]models.Market, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "trending_score", Value: -1}}).
		SetLimit(int64(limit))

	filter := bson.M{"active": true, "closed": false}
	return s.findMarkets(ctx, filter, opts)
}

// GetMarketsByCategory returns markets for a specific category.
func (s *Store) GetMarketsByCategory(ctx context.Context, category string, limit int) ([]models.Market, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "volume_24h", Value: -1}}).
		SetLimit(int64(limit))

	filter := bson.M{"category": category, "active": true, "closed": false}
	return s.findMarkets(ctx, filter, opts)
}

// GetNewMarkets returns recently added markets.
func (s *Store) GetNewMarkets(ctx context.Context, since time.Duration, limit int) ([]models.Market, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "first_seen_at", Value: -1}}).
		SetLimit(int64(limit))

	filter := bson.M{
		"first_seen_at": bson.M{"$gte": time.Now().Add(-since)},
		"active":        true,
		"closed":        false,
	}
	return s.findMarkets(ctx, filter, opts)
}

// GetBreakingMarkets returns markets with significant price movements.
func (s *Store) GetBreakingMarkets(ctx context.Context, threshold float64, limit int) ([]models.Market, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "change_24h", Value: -1}}).
		SetLimit(int64(limit))

	filter := bson.M{
		"$or": []bson.M{
			{"change_24h": bson.M{"$gte": threshold}},
			{"change_24h": bson.M{"$lte": -threshold}},
		},
		"active": true,
		"closed": false,
	}
	return s.findMarkets(ctx, filter, opts)
}

// GetTopMarketsByVolume returns top markets by 24h volume.
func (s *Store) GetTopMarketsByVolume(ctx context.Context, limit int) ([]models.Market, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "volume_24h", Value: -1}}).
		SetLimit(int64(limit))

	filter := bson.M{"active": true, "closed": false}
	return s.findMarkets(ctx, filter, opts)
}

// GetAllActiveMarkets returns all active markets.
func (s *Store) GetAllActiveMarkets(ctx context.Context) ([]models.Market, error) {
	filter := bson.M{"active": true, "closed": false}
	return s.findMarkets(ctx, filter, nil)
}

func (s *Store) findMarkets(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.Market, error) {
	cursor, err := s.markets.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var markets []models.Market
	if err := cursor.All(ctx, &markets); err != nil {
		return nil, err
	}
	return markets, nil
}

// ============================================================================
// SNAPSHOT OPERATIONS
// ============================================================================

// SaveSnapshot saves a market snapshot.
func (s *Store) SaveSnapshot(ctx context.Context, snapshot *models.Snapshot) error {
	snapshot.CapturedAt = time.Now()
	_, err := s.snapshots.InsertOne(ctx, snapshot)
	return err
}

// GetSnapshots returns snapshots for a market within a time range.
func (s *Store) GetSnapshots(ctx context.Context, marketID string, since time.Duration) ([]models.Snapshot, error) {
	filter := bson.M{
		"market_id":   marketID,
		"captured_at": bson.M{"$gte": time.Now().Add(-since)},
	}
	opts := options.Find().SetSort(bson.D{{Key: "captured_at", Value: -1}})

	cursor, err := s.snapshots.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var snapshots []models.Snapshot
	if err := cursor.All(ctx, &snapshots); err != nil {
		return nil, err
	}
	return snapshots, nil
}

// GetLatestSnapshot returns the most recent snapshot for a market.
func (s *Store) GetLatestSnapshot(ctx context.Context, marketID string) (*models.Snapshot, error) {
	var snapshot models.Snapshot
	opts := options.FindOne().SetSort(bson.D{{Key: "captured_at", Value: -1}})
	err := s.snapshots.FindOne(ctx, bson.M{"market_id": marketID}, opts).Decode(&snapshot)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// CleanOldSnapshots removes snapshots older than the given duration.
func (s *Store) CleanOldSnapshots(ctx context.Context, olderThan time.Duration) (int64, error) {
	filter := bson.M{"captured_at": bson.M{"$lt": time.Now().Add(-olderThan)}}
	result, err := s.snapshots.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// ============================================================================
// ARTICLE OPERATIONS
// ============================================================================

// SaveArticle saves a new article.
func (s *Store) SaveArticle(ctx context.Context, article *models.Article) error {
	article.CreatedAt = time.Now()
	article.UpdatedAt = time.Now()
	if article.PublishedAt.IsZero() && article.Published {
		article.PublishedAt = time.Now()
	}

	_, err := s.articles.InsertOne(ctx, article)
	return err
}

// UpdateArticle updates an existing article.
func (s *Store) UpdateArticle(ctx context.Context, article *models.Article) error {
	article.UpdatedAt = time.Now()
	filter := bson.M{"_id": article.ID}
	update := bson.M{"$set": article}
	_, err := s.articles.UpdateOne(ctx, filter, update)
	return err
}

// GetArticleBySlug returns an article by its slug.
func (s *Store) GetArticleBySlug(ctx context.Context, slug string) (*models.Article, error) {
	var article models.Article
	err := s.articles.FindOne(ctx, bson.M{"slug": slug}).Decode(&article)
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// GetArticleByID returns an article by its MongoDB ID.
func (s *Store) GetArticleByID(ctx context.Context, id primitive.ObjectID) (*models.Article, error) {
	var article models.Article
	err := s.articles.FindOne(ctx, bson.M{"_id": id}).Decode(&article)
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// GetRecentArticles returns the most recent published articles.
func (s *Store) GetRecentArticles(ctx context.Context, limit int) ([]models.Article, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "published_at", Value: -1}}).
		SetLimit(int64(limit))

	filter := bson.M{"published": true}
	return s.findArticles(ctx, filter, opts)
}

// GetArticlesByType returns articles of a specific type.
func (s *Store) GetArticlesByType(ctx context.Context, articleType models.ArticleType, limit int) ([]models.Article, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "published_at", Value: -1}}).
		SetLimit(int64(limit))

	filter := bson.M{"type": articleType, "published": true}
	return s.findArticles(ctx, filter, opts)
}

// GetArticlesByCategory returns articles for a specific category.
func (s *Store) GetArticlesByCategory(ctx context.Context, category string, limit int) ([]models.Article, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "published_at", Value: -1}}).
		SetLimit(int64(limit))

	filter := bson.M{"category": category, "published": true}
	return s.findArticles(ctx, filter, opts)
}

// GetFeaturedArticles returns featured articles.
func (s *Store) GetFeaturedArticles(ctx context.Context, limit int) ([]models.Article, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "published_at", Value: -1}}).
		SetLimit(int64(limit))

	filter := bson.M{"featured": true, "published": true}
	return s.findArticles(ctx, filter, opts)
}

// GetTodayArticles returns articles published today.
func (s *Store) GetTodayArticles(ctx context.Context) ([]models.Article, error) {
	today := time.Now().Truncate(24 * time.Hour)
	filter := bson.M{
		"published_at": bson.M{"$gte": today},
		"published":    true,
	}
	opts := options.Find().SetSort(bson.D{{Key: "published_at", Value: -1}})
	return s.findArticles(ctx, filter, opts)
}

// IncrementArticleViews increments the view count for an article.
func (s *Store) IncrementArticleViews(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$inc": bson.M{"views": 1}}
	_, err := s.articles.UpdateOne(ctx, filter, update)
	return err
}

func (s *Store) findArticles(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.Article, error) {
	cursor, err := s.articles.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []models.Article
	if err := cursor.All(ctx, &articles); err != nil {
		return nil, err
	}
	return articles, nil
}

// ============================================================================
// CATEGORY OPERATIONS
// ============================================================================

// GetCategories returns all categories.
func (s *Store) GetCategories(ctx context.Context) ([]models.Category, error) {
	opts := options.Find().SetSort(bson.D{{Key: "order", Value: 1}})
	cursor, err := s.categories.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []models.Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

// GetCategoryBySlug returns a category by its slug.
func (s *Store) GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error) {
	var category models.Category
	err := s.categories.FindOne(ctx, bson.M{"slug": slug}).Decode(&category)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// ============================================================================
// STATS OPERATIONS
// ============================================================================

// Stats holds general statistics.
type Stats struct {
	TotalMarkets   int64 `json:"total_markets"`
	ActiveMarkets  int64 `json:"active_markets"`
	TotalArticles  int64 `json:"total_articles"`
	TodayArticles  int64 `json:"today_articles"`
	TotalSnapshots int64 `json:"total_snapshots"`
}

// GetStats returns general statistics.
func (s *Store) GetStats(ctx context.Context) (*Stats, error) {
	stats := &Stats{}

	var err error
	stats.TotalMarkets, err = s.markets.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	stats.ActiveMarkets, err = s.markets.CountDocuments(ctx, bson.M{"active": true, "closed": false})
	if err != nil {
		return nil, err
	}

	stats.TotalArticles, err = s.articles.CountDocuments(ctx, bson.M{"published": true})
	if err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	stats.TodayArticles, err = s.articles.CountDocuments(ctx, bson.M{
		"published_at": bson.M{"$gte": today},
		"published":    true,
	})
	if err != nil {
		return nil, err
	}

	stats.TotalSnapshots, err = s.snapshots.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	return stats, nil
}

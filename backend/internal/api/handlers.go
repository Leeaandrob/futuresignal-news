package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/leeaandrob/futuresignals/internal/models"
	"github.com/leeaandrob/futuresignals/internal/storage"
)

// Handlers holds the API handlers.
type Handlers struct {
	store *storage.Store
}

// NewHandlers creates new API handlers.
func NewHandlers(store *storage.Store) *Handlers {
	return &Handlers{store: store}
}

// Response helpers

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func getLimit(r *http.Request, defaultLimit int) int {
	limit := defaultLimit
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	return limit
}

// ============================================================================
// ARTICLE HANDLERS
// ============================================================================

// GetArticles returns recent articles.
func (h *Handlers) GetArticles(w http.ResponseWriter, r *http.Request) {
	limit := getLimit(r, 20)

	articles, err := h.store.GetRecentArticles(r.Context(), limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch articles")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"articles": articles,
		"count":    len(articles),
	})
}

// GetArticleBySlug returns a single article by slug.
func (h *Handlers) GetArticleBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		respondError(w, http.StatusBadRequest, "Slug is required")
		return
	}

	article, err := h.store.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		respondError(w, http.StatusNotFound, "Article not found")
		return
	}

	// Increment views
	h.store.IncrementArticleViews(r.Context(), article.ID)

	respondJSON(w, http.StatusOK, article)
}

// GetArticlesByType returns articles of a specific type.
func (h *Handlers) GetArticlesByType(w http.ResponseWriter, r *http.Request) {
	articleType := chi.URLParam(r, "type")
	limit := getLimit(r, 20)

	articles, err := h.store.GetArticlesByType(r.Context(), models.ArticleType(articleType), limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch articles")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"articles": articles,
		"type":     articleType,
		"count":    len(articles),
	})
}

// GetArticlesByCategory returns articles for a category.
func (h *Handlers) GetArticlesByCategory(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	limit := getLimit(r, 20)

	articles, err := h.store.GetArticlesByCategory(r.Context(), category, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch articles")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"articles": articles,
		"category": category,
		"count":    len(articles),
	})
}

// GetBreakingArticles returns breaking news articles.
func (h *Handlers) GetBreakingArticles(w http.ResponseWriter, r *http.Request) {
	limit := getLimit(r, 10)

	articles, err := h.store.GetArticlesByType(r.Context(), models.ArticleTypeBreaking, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch articles")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"articles": articles,
		"count":    len(articles),
	})
}

// GetTrendingArticles returns trending articles.
func (h *Handlers) GetTrendingArticles(w http.ResponseWriter, r *http.Request) {
	limit := getLimit(r, 10)

	articles, err := h.store.GetArticlesByType(r.Context(), models.ArticleTypeTrending, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch articles")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"articles": articles,
		"count":    len(articles),
	})
}

// GetFeaturedArticles returns featured articles.
func (h *Handlers) GetFeaturedArticles(w http.ResponseWriter, r *http.Request) {
	limit := getLimit(r, 5)

	articles, err := h.store.GetFeaturedArticles(r.Context(), limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch articles")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"articles": articles,
		"count":    len(articles),
	})
}

// GetTodayArticles returns articles published today.
func (h *Handlers) GetTodayArticles(w http.ResponseWriter, r *http.Request) {
	articles, err := h.store.GetTodayArticles(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch articles")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"articles": articles,
		"count":    len(articles),
	})
}

// ============================================================================
// MARKET HANDLERS
// ============================================================================

// GetMarkets returns markets.
func (h *Handlers) GetMarkets(w http.ResponseWriter, r *http.Request) {
	limit := getLimit(r, 50)

	markets, err := h.store.GetTopMarketsByVolume(r.Context(), limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch markets")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"markets": markets,
		"count":   len(markets),
	})
}

// GetMarketBySlug returns a single market by slug.
func (h *Handlers) GetMarketBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		respondError(w, http.StatusBadRequest, "Slug is required")
		return
	}

	market, err := h.store.GetMarketBySlug(r.Context(), slug)
	if err != nil {
		respondError(w, http.StatusNotFound, "Market not found")
		return
	}

	respondJSON(w, http.StatusOK, market)
}

// GetTrendingMarkets returns trending markets.
func (h *Handlers) GetTrendingMarkets(w http.ResponseWriter, r *http.Request) {
	limit := getLimit(r, 20)

	markets, err := h.store.GetTrendingMarkets(r.Context(), limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch markets")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"markets": markets,
		"count":   len(markets),
	})
}

// GetMarketsByCategory returns markets for a category.
func (h *Handlers) GetMarketsByCategory(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	limit := getLimit(r, 20)

	markets, err := h.store.GetMarketsByCategory(r.Context(), category, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch markets")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"markets":  markets,
		"category": category,
		"count":    len(markets),
	})
}

// GetNewMarkets returns recently created markets.
func (h *Handlers) GetNewMarkets(w http.ResponseWriter, r *http.Request) {
	limit := getLimit(r, 20)

	markets, err := h.store.GetNewMarkets(r.Context(), 24*7, limit) // Last 7 days
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch markets")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"markets": markets,
		"count":   len(markets),
	})
}

// GetBreakingMarkets returns markets with significant movements.
func (h *Handlers) GetBreakingMarkets(w http.ResponseWriter, r *http.Request) {
	limit := getLimit(r, 20)

	markets, err := h.store.GetBreakingMarkets(r.Context(), 0.05, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch markets")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"markets": markets,
		"count":   len(markets),
	})
}

// ============================================================================
// CATEGORY HANDLERS
// ============================================================================

// GetCategories returns all categories.
func (h *Handlers) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.store.GetCategories(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
		"count":      len(categories),
	})
}

// GetCategoryBySlug returns a single category with its content.
func (h *Handlers) GetCategoryBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	category, err := h.store.GetCategoryBySlug(r.Context(), slug)
	if err != nil {
		respondError(w, http.StatusNotFound, "Category not found")
		return
	}

	// Get markets and articles for this category
	markets, _ := h.store.GetMarketsByCategory(r.Context(), slug, 10)
	articles, _ := h.store.GetArticlesByCategory(r.Context(), slug, 10)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"category": category,
		"markets":  markets,
		"articles": articles,
	})
}

// ============================================================================
// STATS HANDLERS
// ============================================================================

// GetStats returns general statistics.
func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.GetStats(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch stats")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// HealthCheck returns service health.
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "futuresignals",
	})
}

// ============================================================================
// FEED HANDLERS (for homepage)
// ============================================================================

// GetHomeFeed returns curated content for the homepage.
func (h *Handlers) GetHomeFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get featured/breaking articles
	featured, _ := h.store.GetFeaturedArticles(ctx, 3)
	if len(featured) == 0 {
		featured, _ = h.store.GetArticlesByType(ctx, models.ArticleTypeBreaking, 3)
	}

	// Get recent articles
	recent, _ := h.store.GetRecentArticles(ctx, 10)

	// Get trending markets
	trendingMarkets, _ := h.store.GetTrendingMarkets(ctx, 10)

	// Get today's briefings
	todayArticles, _ := h.store.GetTodayArticles(ctx)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"featured":         featured,
		"recent":           recent,
		"trending_markets": trendingMarkets,
		"today":            todayArticles,
	})
}

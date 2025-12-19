package api

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/leeaandrob/futuresignals/internal/scheduler"
	"github.com/leeaandrob/futuresignals/internal/storage"
	syncer "github.com/leeaandrob/futuresignals/internal/sync"
	"github.com/rs/zerolog/log"
)

// Server represents the API server.
type Server struct {
	router    *chi.Mux
	handlers  *Handlers
	syncer    *syncer.Syncer
	scheduler *scheduler.Scheduler
	addr      string
	server    *http.Server
}

// NewServer creates a new API server.
func NewServer(store *storage.Store, s *syncer.Syncer, sched *scheduler.Scheduler, addr string) *Server {
	handlers := NewHandlers(store)

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Routes
	r.Route("/api", func(r chi.Router) {
		// Health
		r.Get("/health", handlers.HealthCheck)
		r.Get("/stats", handlers.GetStats)

		// Home feed
		r.Get("/feed", handlers.GetHomeFeed)

		// Articles
		r.Route("/articles", func(r chi.Router) {
			r.Get("/", handlers.GetArticles)
			r.Get("/today", handlers.GetTodayArticles)
			r.Get("/breaking", handlers.GetBreakingArticles)
			r.Get("/trending", handlers.GetTrendingArticles)
			r.Get("/featured", handlers.GetFeaturedArticles)
			r.Get("/type/{type}", handlers.GetArticlesByType)
			r.Get("/category/{category}", handlers.GetArticlesByCategory)
			r.Get("/{slug}", handlers.GetArticleBySlug)
		})

		// Markets
		r.Route("/markets", func(r chi.Router) {
			r.Get("/", handlers.GetMarkets)
			r.Get("/trending", handlers.GetTrendingMarkets)
			r.Get("/breaking", handlers.GetBreakingMarkets)
			r.Get("/new", handlers.GetNewMarkets)
			r.Get("/category/{category}", handlers.GetMarketsByCategory)
			r.Get("/{slug}", handlers.GetMarketBySlug)
		})

		// Categories
		r.Route("/categories", func(r chi.Router) {
			r.Get("/", handlers.GetCategories)
			r.Get("/{slug}", handlers.GetCategoryBySlug)
		})
	})

	// Create server instance for admin routes closure
	srv := &Server{
		router:    r,
		handlers:  handlers,
		syncer:    s,
		scheduler: sched,
		addr:      addr,
	}

	// Admin routes (no auth for development)
	r.Route("/api/admin", func(r chi.Router) {
		// Force sync markets
		r.Post("/sync", srv.AdminSyncNow)
		r.Get("/debug", srv.AdminDebugSync)

		// Job management
		r.Get("/jobs", srv.AdminGetJobs)
		r.Post("/jobs/{name}/run", srv.AdminRunJob)
	})

	return srv
}

// Start starts the API server.
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info().Str("addr", s.addr).Msg("Starting API server")
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// ============================================================================
// ADMIN HANDLERS
// ============================================================================

// AdminSyncNow forces an immediate market sync.
func (s *Server) AdminSyncNow(w http.ResponseWriter, r *http.Request) {
	if s.syncer == nil {
		respondError(w, http.StatusServiceUnavailable, "Syncer not available")
		return
	}

	go s.syncer.SyncNow()

	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Sync triggered",
	})
}

// AdminGetJobs returns the status of all scheduled jobs.
func (s *Server) AdminGetJobs(w http.ResponseWriter, r *http.Request) {
	if s.scheduler == nil {
		respondError(w, http.StatusServiceUnavailable, "Scheduler not available")
		return
	}

	jobs := s.scheduler.GetJobStatus()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"jobs":  jobs,
		"count": len(jobs),
	})
}

// AdminDebugSync fetches markets from Polymarket and returns debug info.
func (s *Server) AdminDebugSync(w http.ResponseWriter, r *http.Request) {
	if s.syncer == nil {
		respondError(w, http.StatusServiceUnavailable, "Syncer not available")
		return
	}

	// Get cached markets from syncer
	markets := s.syncer.GetTrendingMarkets(20)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"cached_market_count": len(markets),
		"markets":             markets,
	})
}

// AdminRunJob runs a specific job by name.
func (s *Server) AdminRunJob(w http.ResponseWriter, r *http.Request) {
	if s.scheduler == nil {
		respondError(w, http.StatusServiceUnavailable, "Scheduler not available")
		return
	}

	name := chi.URLParam(r, "name")
	if name == "" {
		respondError(w, http.StatusBadRequest, "Job name is required")
		return
	}

	if err := s.scheduler.RunJobNow(name); err != nil {
		respondError(w, http.StatusNotFound, "Job not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Job triggered: " + name,
	})
}

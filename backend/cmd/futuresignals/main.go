// FutureSignals - The Bloomberg of Prediction Markets
// Monitors prediction markets and generates editorial content.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/leeaandrob/futuresignals/internal/api"
	"github.com/leeaandrob/futuresignals/internal/config"
	"github.com/leeaandrob/futuresignals/internal/content"
	"github.com/leeaandrob/futuresignals/internal/enrichment"
	"github.com/leeaandrob/futuresignals/internal/polymarket"
	"github.com/leeaandrob/futuresignals/internal/qwen"
	"github.com/leeaandrob/futuresignals/internal/scheduler"
	"github.com/leeaandrob/futuresignals/internal/storage"
	syncer "github.com/leeaandrob/futuresignals/internal/sync"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("FutureSignals - Starting content engine")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid configuration")
	}

	ctx := context.Background()

	// Initialize storage
	store, err := storage.NewStore(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer store.Close(ctx)

	// Initialize Polymarket client
	pmClient := polymarket.NewClient()
	log.Info().Msg("Polymarket client initialized")

	// Initialize Qwen LLM client
	var llmClient *qwen.Client
	if cfg.DashScopeAPIKey != "" {
		llmClient = qwen.NewClient(qwen.Config{
			APIKey:   cfg.DashScopeAPIKey,
			Endpoint: cfg.DashScopeEndpoint,
			Model:    cfg.QwenModel,
		})
		log.Info().Str("model", cfg.QwenModel).Msg("Qwen LLM client initialized")
	} else {
		log.Warn().Msg("Qwen client not initialized (no API key)")
	}

	// Initialize enrichment pipeline
	var enricher *enrichment.Enricher
	if cfg.EnableEnrichment {
		enricher = enrichment.NewEnricher(enrichment.EnrichmentConfig{
			TavilyAPIKey:    cfg.TavilyAPIKey,
			ExaAPIKey:       cfg.ExaAPIKey,
			FirecrawlAPIKey: cfg.FirecrawlAPIKey,
			MaxNewsResults:  5,
			MaxDeepScrapes:  2,
			EnableTavily:    cfg.TavilyAPIKey != "",
			EnableExa:       cfg.ExaAPIKey != "",
			EnableFirecrawl: cfg.FirecrawlAPIKey != "",
		})
		log.Info().Msg("Enrichment pipeline initialized")
	}

	// Initialize market syncer
	syncConfig := syncer.DefaultSyncerConfig()
	syncConfig.SyncInterval = cfg.PollInterval
	syncConfig.MinVolume24h = cfg.MinVolume24h
	syncConfig.BreakingThreshold = cfg.MinProbabilityChange

	marketSyncer := syncer.NewSyncer(pmClient, store, syncConfig)
	log.Info().Msg("Market syncer initialized")

	// Initialize content generator
	generator := content.NewGenerator(store, marketSyncer, llmClient, enricher)
	log.Info().Msg("Content generator initialized")

	// Initialize scheduler
	sched := scheduler.NewScheduler(generator, marketSyncer)
	log.Info().Msg("Scheduler initialized")

	// Initialize API server with syncer and scheduler for admin endpoints
	apiServer := api.NewServer(store, marketSyncer, sched, cfg.HTTPAddr)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start all services
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Error().Err(err).Msg("API server error")
		}
	}()

	marketSyncer.Start()
	sched.Start()

	log.Info().
		Str("api", cfg.HTTPAddr).
		Msg("FutureSignals engine running")

	// Wait for shutdown signal
	<-sigChan
	log.Info().Msg("Shutdown signal received")

	// Graceful shutdown
	shutdownCtx := context.Background()
	sched.Stop()
	marketSyncer.Stop()
	apiServer.Shutdown(shutdownCtx)

	log.Info().Msg("FutureSignals engine stopped")
}

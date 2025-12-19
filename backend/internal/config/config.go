// Package config provides configuration management for FutureSignals.
package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// Config holds all application configuration.
type Config struct {
	// Qwen/DashScope settings
	DashScopeAPIKey   string
	DashScopeEndpoint string
	QwenModel         string

	// Enrichment API settings
	TavilyAPIKey    string
	ExaAPIKey       string
	FirecrawlAPIKey string
	EnableEnrichment bool

	// MongoDB settings
	MongoURI string
	MongoDB  string

	// Detector settings
	MinProbabilityChange float64
	MinVolume24h         float64
	PollInterval         time.Duration

	// Server settings
	HTTPAddr string
	Debug    bool
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	// Try to load .env file
	if err := godotenv.Load(); err != nil {
		log.Debug().Msg("No .env file found, using environment variables")
	}

	cfg := &Config{
		// Qwen/DashScope
		DashScopeAPIKey:   getEnv("DASHSCOPE_API_KEY", ""),
		DashScopeEndpoint: getEnv("DASHSCOPE_ENDPOINT", "https://dashscope-intl.aliyuncs.com/compatible-mode/v1"),
		QwenModel:         getEnv("QWEN_MODEL", "qwen-plus"),

		// Enrichment APIs
		TavilyAPIKey:     getEnv("TAVILY_API_KEY", ""),
		ExaAPIKey:        getEnv("EXA_API_KEY", ""),
		FirecrawlAPIKey:  getEnv("FIRECRAWL_API_KEY", ""),
		EnableEnrichment: getEnvBool("ENABLE_ENRICHMENT", true),

		// MongoDB
		MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:  getEnv("MONGO_DB", "futuresignals"),

		// Detector
		MinProbabilityChange: getEnvFloat("MIN_PROBABILITY_CHANGE", 0.07),
		MinVolume24h:         getEnvFloat("MIN_VOLUME_24H", 50000),
		PollInterval:         getEnvDuration("POLL_INTERVAL", 5*time.Minute),

		// Server
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),
		Debug:    getEnvBool("DEBUG", false),
	}

	return cfg, nil
}

// Validate checks if required configuration is present.
func (c *Config) Validate() error {
	if c.DashScopeAPIKey == "" {
		log.Warn().Msg("DASHSCOPE_API_KEY not set, narrative generation will be disabled")
	}
	return nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

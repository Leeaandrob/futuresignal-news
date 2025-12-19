# Changelog

All notable changes to FutureSignals will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- RSS/Atom feeds for content syndication
- Email newsletter integration (Resend)
- Push notifications for breaking signals
- Audio briefings via ElevenLabs TTS
- Support for Kalshi and Metaculus prediction markets
- Premium API access tier

---

## [1.1.0] - 2024-12-19

### Added
- **XTracker Integration**: Social signal correlation with Polymarket's Twitter tracker
  - New `xtracker/` package with API client and correlator
  - Fetches posts from tracked influencers (currently @elonmusk)
  - Correlates tweets with market movements via keyword matching
  - 2-hour impact window for before/after probability analysis

- **Social Signals in Articles**
  - `SocialSignal` and `MarketMovement` types in `models/article.go`
  - `socialSignals` field added to Article model
  - Articles enriched with correlated social signals during generation

- **Frontend Social Signal Components**
  - `SocialSignalCard.tsx` with default, compact, and inline variants
  - `SocialSignalsSection` for full signal display
  - `SignalSourceBadges` for article header influencer badges
  - Tweet content, engagement metrics, and market impact display

- **LLM Prompt Enhancement**
  - `SocialSignalsContext` field in `qwen.SignalData`
  - Updated Bloomberg-style prompts to cite influencers as sources
  - Social signals section in narrative generation prompts

- **Social Sharing**
  - Twitter share button on article pages
  - Copy link button with clipboard API

### Changed
- Moved `SocialSignal` and `MarketMovement` types from xtracker to models package (circular dependency fix)
- Updated article page (`[slug].astro`) with social signals sections
- Content generator now calls `enrichWithSocialSignals()` for all article types

### Technical
- Backend: v1.1.0 Docker image deployed to Kubernetes
- Frontend: Auto-deployed via Cloudflare Pages

---

## [1.0.1] - 2024-12-19

### Fixed
- Probability display formatting in market cards
- Event volume tracking and display

### Changed
- Optimized backfill to fetch markets by ID
- Use event slug for correct Polymarket URLs

---

## [1.0.0] - 2024-12-18

### Added
- **Backend Core**
  - Go 1.23+ with Gin HTTP framework
  - MongoDB Atlas integration for persistence
  - Polymarket API client with rate limiting
  - Signal detection with configurable thresholds

- **Content Generation**
  - Qwen LLM (DashScope) integration for narratives
  - Bloomberg-style editorial prompts
  - Perplexity API for external context enrichment
  - Multiple article types: breaking, trending, briefing, new_market, deep_dive

- **REST API**
  - `/api/articles` - Article CRUD with pagination
  - `/api/markets` - Market listing and details
  - `/api/categories` - Category management
  - `/api/sentiment` - Market Pulse momentum data
  - `/api/feed/home` - Homepage feed aggregation
  - `/api/stats` - Platform statistics
  - `/health` - Service health check

- **Frontend**
  - Astro 5.x with React islands architecture
  - shadcn/ui component library
  - Tailwind CSS 4 styling
  - SSR via Cloudflare Workers
  - Market cards with probability bars
  - Article pages with SEO optimization
  - JSON-LD schema for Google News
  - Market Pulse sentiment header bar
  - Category pages with filtering

- **Infrastructure**
  - Docker containerization
  - Kubernetes deployment manifests
  - GHCR image registry
  - Cloudflare Pages for frontend
  - MongoDB Atlas M0 cluster

- **Tools**
  - Backfill command for historical data import
  - Signal daemon for continuous monitoring

### Technical Details
- Signal detection: >=5% probability change, >$10k volume
- Polling interval: 5 minutes default
- Categories: politics, crypto, tech, sports, finance, world, culture
- Article expiration: 7 days default

---

## Release Links

[Unreleased]: https://github.com/leeaandrob/futuresignals/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/leeaandrob/futuresignals/compare/v1.0.1...v1.1.0
[1.0.1]: https://github.com/leeaandrob/futuresignals/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/leeaandrob/futuresignals/releases/tag/v1.0.0

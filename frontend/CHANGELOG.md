# Changelog

All notable changes to FutureSignals will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-12-19

### Added
- **SSR with Cloudflare Workers**: Article and market pages now render server-side for instant content updates
- **Google Analytics**: GA4 tracking (G-EB976WFQ5J) integrated
- **Google AdSense**: Ad network integration ready
- **Comprehensive SEO**:
  - JSON-LD NewsArticle schema for Google News compatibility
  - JSON-LD NewsMediaOrganization schema
  - Complete Open Graph tags with absolute image URLs
  - Twitter Cards with summary_large_image
  - robots.txt with sitemap and Google News directives
  - news_keywords meta tag
  - article:published_time, section, and tags meta
- **RSS Feed**: Full RSS 2.0 feed at /rss.xml
- **Sitemap**: Auto-generated sitemap-index.xml
- **404 Page**: Custom error page
- **Categories**: Trending, Breaking, Politics, Crypto, Tech, Sports, Finance, World, Economy, Elections, Culture

### Infrastructure
- Astro 5 with React integration
- TailwindCSS for styling
- Cloudflare Pages deployment with Workers for SSR
- Backend API at api.futuresignals.news
- MongoDB Atlas for data storage

### Fixed
- Fixed NaN% display in market probability
- Fixed Invalid Date formatting
- Fixed undefined errors in MarketCard component
- Fixed RSS feed site URL

---

## [Unreleased]

### Planned
- Dynamic OG images per article
- Push notifications for breaking news
- User accounts and watchlists
- Dark mode toggle

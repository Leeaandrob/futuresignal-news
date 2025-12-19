// Types matching the new backend models

// =============================================================================
// CATEGORIES
// =============================================================================

export interface Category {
  slug: string;
  name: string;
  description: string;
  icon: string;
  color: string;
  dynamic: boolean;
}

// Category slugs matching Polymarket
export type CategorySlug =
  | "trending"
  | "breaking"
  | "new"
  | "politics"
  | "elections"
  | "crypto"
  | "finance"
  | "economy"
  | "earnings"
  | "tech"
  | "sports"
  | "geopolitics"
  | "world"
  | "culture";

// =============================================================================
// MARKETS
// =============================================================================

export interface PolymarketTag {
  label: string;
  slug: string;
}

export interface Market {
  id: string;
  marketId: string;
  conditionId: string;
  groupItemTitle: string;
  slug: string;
  question: string;
  description: string;
  category: string;

  // Media
  image?: string;
  icon?: string;

  // Pricing
  probability: number;
  previousProb: number;
  lastTradePrice?: number;
  change1h?: number;
  change24h: number;
  change7d?: number;

  // Volume
  volume1h?: number;
  volume24h: number;
  volume7d?: number;
  totalVolume: number;

  // Event-level data
  eventVolume?: number;
  eventVolume24h?: number;
  eventTitle?: string;

  // Engagement
  commentCount?: number;
  competitorCount?: number;

  // Classification
  polymarketTags?: PolymarketTag[];

  // Resolution
  resolutionSource?: string;
  seriesSlug?: string;
  startDate?: string;

  // Status
  liquidity: number;
  active: boolean;
  closed: boolean;
  archived: boolean;
  acceptingBid: boolean;
  endDate: string;

  // Outcomes
  outcomes: string[];
  outcomePrices: number[];

  // Meta
  trendingScore: number;
  firstSeenAt: string;
  updatedAt: string;
  polymarketUrl: string;
}

export interface Snapshot {
  id: string;
  marketId: string;
  probability: number;
  volume24h: number;
  totalVolume: number;
  liquidity: number;
  timestamp: string;
}

// =============================================================================
// ARTICLES
// =============================================================================

export type ArticleType =
  | "breaking"
  | "briefing"
  | "trending"
  | "new_market"
  | "deep_dive"
  | "digest"
  | "explainer";

export type BriefingType = "morning" | "midday" | "evening" | "weekly";

export interface ArticleBody {
  whatHappened: string;
  whyItMatters: string;
  context: string[];
  whatToWatch: string;
}

export interface RelatedArticle {
  slug: string;
  title: string;
  type: ArticleType;
}

export interface Article {
  id: string;
  slug: string;
  type: ArticleType;
  title: string;
  subtitle: string;
  summary: string;
  body: ArticleBody;
  category: string;
  markets: Market[];
  marketIds: string[];
  relatedArticles: RelatedArticle[];
  tags: string[];
  featured: boolean;
  priority: number;
  views: number;
  publishedAt: string;
  updatedAt: string;
  expiresAt: string;
  briefingType?: BriefingType;
  enrichmentSources?: string[];
}

// =============================================================================
// API RESPONSES
// =============================================================================

export interface ArticlesResponse {
  articles: Article[];
  count: number;
}

export interface MarketsResponse {
  markets: Market[];
  count: number;
}

export interface CategoriesResponse {
  categories: Category[];
  count: number;
}

export interface CategoryDetailResponse {
  category: Category;
  markets: Market[];
  articles: Article[];
}

export interface HomeFeedResponse {
  featured: Article[];
  recent: Article[];
  trendingMarkets: Market[];
  today: Article[];
}

export interface StatsResponse {
  totalMarkets: number;
  activeMarkets: number;
  totalArticles: number;
  articlesToday: number;
  totalSnapshots: number;
  categories: number;
}

export interface HealthResponse {
  status: string;
  service: string;
}

// =============================================================================
// SENTIMENT / MARKET PULSE
// =============================================================================

export interface CategorySentiment {
  category: string;
  name: string;
  color: string;
  icon: string;
  momentum: number;           // Volume-weighted avg change (-1 to 1)
  totalVolume24h: number;     // Sum of all volume24h
  marketCount: number;        // Active markets count
  breakingCount: number;      // Markets with |change| > 10%
  topMover?: string;          // Market with highest |change|
  topMoverSlug?: string;      // Slug for link
  topMoverChange: number;     // Change of top mover
  avgChange24h: number;       // Simple average change
}

export interface SentimentResponse {
  sentiments: CategorySentiment[];
  count: number;
}

// =============================================================================
// UI HELPERS
// =============================================================================

export interface BreadcrumbItem {
  label: string;
  href?: string;
}

export type SignificanceLevel = "low" | "medium" | "high" | "breaking";

export function getSignificance(change: number): SignificanceLevel {
  const abs = Math.abs(change);
  if (abs >= 0.15) return "breaking";
  if (abs >= 0.10) return "high";
  if (abs >= 0.05) return "medium";
  return "low";
}

export function getArticleTypeBadge(type: ArticleType): { label: string; variant: string } {
  const badges: Record<ArticleType, { label: string; variant: string }> = {
    breaking: { label: "BREAKING", variant: "breaking" },
    briefing: { label: "BRIEFING", variant: "default" },
    trending: { label: "TRENDING", variant: "bullish" },
    new_market: { label: "NEW", variant: "tech" },
    deep_dive: { label: "ANALYSIS", variant: "politics" },
    digest: { label: "DIGEST", variant: "secondary" },
    explainer: { label: "EXPLAINER", variant: "secondary" },
  };
  return badges[type] || { label: type.toUpperCase(), variant: "secondary" };
}

// Default categories list for navigation
export const DEFAULT_CATEGORIES: Category[] = [
  { slug: "trending", name: "Trending", description: "Hot markets right now", icon: "trending_up", color: "#FF6B6B", dynamic: true },
  { slug: "breaking", name: "Breaking", description: "Big moves happening now", icon: "bolt", color: "#FF4757", dynamic: true },
  { slug: "politics", name: "Politics", description: "Elections, policy, governance", icon: "account_balance", color: "#6B46C1", dynamic: false },
  { slug: "crypto", name: "Crypto", description: "Bitcoin, Ethereum, DeFi", icon: "currency_bitcoin", color: "#F7931A", dynamic: false },
  { slug: "tech", name: "Tech", description: "AI, companies, innovation", icon: "computer", color: "#0891B2", dynamic: false },
  { slug: "sports", name: "Sports", description: "Games, championships, athletes", icon: "sports_soccer", color: "#10B981", dynamic: false },
  { slug: "finance", name: "Finance", description: "Markets, rates, economy", icon: "attach_money", color: "#3B82F6", dynamic: false },
  { slug: "world", name: "World", description: "Global events, geopolitics", icon: "public", color: "#8B5CF6", dynamic: false },
];

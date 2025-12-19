// API client for FutureSignals backend
import type {
  Article,
  ArticlesResponse,
  Market,
  MarketsResponse,
  Category,
  CategoriesResponse,
  CategoryDetailResponse,
  HomeFeedResponse,
  StatsResponse,
  HealthResponse,
  ArticleType,
} from "./types";

const API_BASE = import.meta.env.PUBLIC_API_URL || "http://localhost:8080";

// =============================================================================
// SNAKE_CASE TO CAMELCASE TRANSFORMER
// =============================================================================

function snakeToCamel(str: string): string {
  return str.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());
}

function transformKeys(obj: any): any {
  if (obj === null || obj === undefined) return obj;
  if (Array.isArray(obj)) return obj.map(transformKeys);
  if (typeof obj !== "object") return obj;

  const transformed: any = {};
  for (const key of Object.keys(obj)) {
    const camelKey = snakeToCamel(key);
    transformed[camelKey] = transformKeys(obj[key]);
  }
  return transformed;
}

// =============================================================================
// GENERIC FETCH HELPER
// =============================================================================

async function apiFetch<T>(endpoint: string): Promise<T> {
  const res = await fetch(`${API_BASE}${endpoint}`);
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`);
  }
  const data = await res.json();
  return transformKeys(data) as T;
}

// =============================================================================
// HOME FEED
// =============================================================================

export async function getHomeFeed(): Promise<HomeFeedResponse> {
  return apiFetch<HomeFeedResponse>("/api/feed");
}

// =============================================================================
// ARTICLES
// =============================================================================

export async function getArticles(limit: number = 20): Promise<Article[]> {
  const data = await apiFetch<ArticlesResponse>(`/api/articles?limit=${limit}`);
  return data.articles || [];
}

export async function getArticleBySlug(slug: string): Promise<Article | null> {
  try {
    return await apiFetch<Article>(`/api/articles/${slug}`);
  } catch {
    return null;
  }
}

export async function getTodayArticles(): Promise<Article[]> {
  const data = await apiFetch<ArticlesResponse>("/api/articles/today");
  return data.articles || [];
}

export async function getBreakingArticles(limit: number = 10): Promise<Article[]> {
  const data = await apiFetch<ArticlesResponse>(`/api/articles/breaking?limit=${limit}`);
  return data.articles || [];
}

export async function getTrendingArticles(limit: number = 10): Promise<Article[]> {
  const data = await apiFetch<ArticlesResponse>(`/api/articles/trending?limit=${limit}`);
  return data.articles || [];
}

export async function getFeaturedArticles(limit: number = 5): Promise<Article[]> {
  const data = await apiFetch<ArticlesResponse>(`/api/articles/featured?limit=${limit}`);
  return data.articles || [];
}

export async function getArticlesByType(type: ArticleType, limit: number = 20): Promise<Article[]> {
  const data = await apiFetch<ArticlesResponse>(`/api/articles/type/${type}?limit=${limit}`);
  return data.articles || [];
}

export async function getArticlesByCategory(category: string, limit: number = 20): Promise<Article[]> {
  const data = await apiFetch<ArticlesResponse>(`/api/articles/category/${category}?limit=${limit}`);
  return data.articles || [];
}

// =============================================================================
// MARKETS
// =============================================================================

export async function getMarkets(limit: number = 50): Promise<Market[]> {
  const data = await apiFetch<MarketsResponse>(`/api/markets?limit=${limit}`);
  return data.markets || [];
}

export async function getMarketBySlug(slug: string): Promise<Market | null> {
  try {
    return await apiFetch<Market>(`/api/markets/${slug}`);
  } catch {
    return null;
  }
}

export async function getTrendingMarkets(limit: number = 20): Promise<Market[]> {
  const data = await apiFetch<MarketsResponse>(`/api/markets/trending?limit=${limit}`);
  return data.markets || [];
}

export async function getBreakingMarkets(limit: number = 20): Promise<Market[]> {
  const data = await apiFetch<MarketsResponse>(`/api/markets/breaking?limit=${limit}`);
  return data.markets || [];
}

export async function getNewMarkets(limit: number = 20): Promise<Market[]> {
  const data = await apiFetch<MarketsResponse>(`/api/markets/new?limit=${limit}`);
  return data.markets || [];
}

export async function getMarketsByCategory(category: string, limit: number = 20): Promise<Market[]> {
  const data = await apiFetch<MarketsResponse>(`/api/markets/category/${category}?limit=${limit}`);
  return data.markets || [];
}

// =============================================================================
// CATEGORIES
// =============================================================================

export async function getCategories(): Promise<Category[]> {
  const data = await apiFetch<CategoriesResponse>("/api/categories");
  return data.categories || [];
}

export async function getCategoryBySlug(slug: string): Promise<CategoryDetailResponse | null> {
  try {
    return await apiFetch<CategoryDetailResponse>(`/api/categories/${slug}`);
  } catch {
    return null;
  }
}

// =============================================================================
// STATS & HEALTH
// =============================================================================

export async function getStats(): Promise<StatsResponse> {
  return apiFetch<StatsResponse>("/api/stats");
}

export async function getHealth(): Promise<HealthResponse> {
  return apiFetch<HealthResponse>("/api/health");
}

// =============================================================================
// RE-EXPORTS
// =============================================================================

export type {
  Article,
  Market,
  Category,
  ArticleType,
  BriefingType,
  ArticleBody,
  HomeFeedResponse,
} from "./types";

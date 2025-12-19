import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// =============================================================================
// NUMBER FORMATTING
// =============================================================================

export function formatVolume(value: number | undefined | null): string {
  if (value == null || isNaN(value)) return "$0";
  if (value >= 1_000_000) {
    return `$${(value / 1_000_000).toFixed(1)}M`;
  }
  if (value >= 1_000) {
    return `$${(value / 1_000).toFixed(0)}K`;
  }
  return `$${value.toFixed(0)}`;
}

export function formatChange(change: number | undefined | null): string {
  if (change == null || isNaN(change)) return "+0.0%";
  const sign = change >= 0 ? "+" : "";
  return `${sign}${(change * 100).toFixed(1)}%`;
}

export function formatProbability(prob: number | undefined | null): string {
  if (prob == null || isNaN(prob)) return "0%";
  return `${(prob * 100).toFixed(0)}%`;
}

export function formatNumber(value: number): string {
  return new Intl.NumberFormat("en-US").format(value);
}

// =============================================================================
// DATE FORMATTING
// =============================================================================

export function formatTimeAgo(date: Date | string | undefined | null): string {
  if (!date) return "";
  const now = new Date();
  const past = new Date(date);
  if (isNaN(past.getTime())) return "";

  const diffMs = now.getTime() - past.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 1) return "Just now";
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;

  return past.toLocaleDateString("en-US", { month: "short", day: "numeric" });
}

export function formatDate(date: Date | string | undefined | null): string {
  if (!date) return "";
  const d = new Date(date);
  if (isNaN(d.getTime())) return "";
  return d.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export function formatDateTime(date: Date | string | undefined | null): string {
  if (!date) return "";
  const d = new Date(date);
  if (isNaN(d.getTime())) return "";
  return d.toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

export function formatFullDate(date: Date | string | undefined | null): string {
  if (!date) return "";
  const d = new Date(date);
  if (isNaN(d.getTime())) return "";
  return d.toLocaleDateString("en-US", {
    weekday: "long",
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

// =============================================================================
// CATEGORY HELPERS
// =============================================================================

export const CATEGORY_COLORS: Record<string, string> = {
  politics: "#6B46C1",
  elections: "#7C3AED",
  crypto: "#F7931A",
  finance: "#3B82F6",
  economy: "#0EA5E9",
  earnings: "#6366F1",
  tech: "#0891B2",
  sports: "#10B981",
  geopolitics: "#8B5CF6",
  world: "#EC4899",
  culture: "#F43F5E",
  trending: "#FF6B6B",
  breaking: "#FF4757",
  new: "#22C55E",
};

export function getCategoryColor(category: string): string {
  return CATEGORY_COLORS[category.toLowerCase()] || "#6B7280";
}

export function getCategoryBadgeClass(category: string): string {
  const cat = category.toLowerCase();
  const colorMap: Record<string, string> = {
    politics: "badge-politics",
    tech: "badge-tech",
    crypto: "badge-crypto",
    sports: "badge-sports",
    finance: "badge-finance",
    world: "badge-world",
    breaking: "badge-breaking",
    trending: "badge-trending",
  };
  return colorMap[cat] || "badge-default";
}

// =============================================================================
// SIGNIFICANCE HELPERS
// =============================================================================

export type SignificanceLevel = "low" | "medium" | "high" | "breaking";

export function getSignificance(change: number): SignificanceLevel {
  const abs = Math.abs(change);
  if (abs >= 0.15) return "breaking";
  if (abs >= 0.10) return "high";
  if (abs >= 0.05) return "medium";
  return "low";
}

export function getSignificanceLabel(significance: SignificanceLevel): string {
  const labels: Record<SignificanceLevel, string> = {
    low: "Minor",
    medium: "Notable",
    high: "Significant",
    breaking: "Breaking",
  };
  return labels[significance];
}

// =============================================================================
// TEXT HELPERS
// =============================================================================

export function truncate(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength).trim() + "...";
}

export function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^\w\s-]/g, "")
    .replace(/[\s_-]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

// =============================================================================
// URL HELPERS
// =============================================================================

export function getArticleUrl(slug: string): string {
  return `/article/${slug}`;
}

export function getMarketUrl(slug: string): string {
  return `/market/${slug}`;
}

export function getCategoryUrl(slug: string): string {
  return `/category/${slug}`;
}

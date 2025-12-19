import { useEffect, useState } from "react";
import { getMarkets, getTrendingMarkets, type Market } from "@/lib/api";
import { MarketCard, MarketGrid, MarketList } from "@/components/MarketCard";
import { Badge } from "@/components/ui/badge";
import { Loader2 } from "lucide-react";

type ViewMode = "grid" | "list";
type SortMode = "volume" | "trending" | "change";

export function MarketsPage() {
  const [markets, setMarkets] = useState<Market[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<ViewMode>("grid");
  const [sortMode, setSortMode] = useState<SortMode>("volume");
  const [categoryFilter, setCategoryFilter] = useState<string | null>(null);

  useEffect(() => {
    async function loadMarkets() {
      setLoading(true);
      setError(null);

      try {
        let data: Market[];
        if (sortMode === "trending") {
          data = await getTrendingMarkets(100);
        } else {
          data = await getMarkets(100);
        }

        // Sort by change if needed
        if (sortMode === "change") {
          data = [...data].sort((a, b) => Math.abs(b.change24h) - Math.abs(a.change24h));
        }

        setMarkets(data);
      } catch (err) {
        setError("Failed to load markets. Please try again.");
        console.error(err);
      } finally {
        setLoading(false);
      }
    }

    loadMarkets();
  }, [sortMode]);

  // Get unique categories from markets
  const categories = [...new Set(markets.map((m) => m.category))].filter(Boolean);

  // Filter markets by category
  const filteredMarkets = categoryFilter
    ? markets.filter((m) => m.category === categoryFilter)
    : markets;

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="w-8 h-8 animate-spin text-brand" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-20">
        <p className="text-muted-foreground">{error}</p>
        <button
          onClick={() => window.location.reload()}
          className="mt-4 text-brand hover:underline"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div>
      {/* Filters and Controls */}
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between mb-8">
        {/* Sort Options */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">Sort by:</span>
          <div className="flex gap-1">
            {(["volume", "trending", "change"] as SortMode[]).map((mode) => (
              <button
                key={mode}
                onClick={() => setSortMode(mode)}
                className={`px-3 py-1.5 text-sm rounded-md transition-colors ${
                  sortMode === mode
                    ? "bg-brand text-white"
                    : "bg-muted text-muted-foreground hover:text-foreground"
                }`}
              >
                {mode.charAt(0).toUpperCase() + mode.slice(1)}
              </button>
            ))}
          </div>
        </div>

        {/* View Toggle */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">View:</span>
          <div className="flex gap-1">
            <button
              onClick={() => setViewMode("grid")}
              className={`p-2 rounded-md transition-colors ${
                viewMode === "grid"
                  ? "bg-brand text-white"
                  : "bg-muted text-muted-foreground hover:text-foreground"
              }`}
              aria-label="Grid view"
            >
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
              </svg>
            </button>
            <button
              onClick={() => setViewMode("list")}
              className={`p-2 rounded-md transition-colors ${
                viewMode === "list"
                  ? "bg-brand text-white"
                  : "bg-muted text-muted-foreground hover:text-foreground"
              }`}
              aria-label="List view"
            >
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      {/* Category Filter */}
      <div className="flex flex-wrap gap-2 mb-6">
        <button
          onClick={() => setCategoryFilter(null)}
          className={`px-3 py-1 text-sm rounded-full transition-colors ${
            categoryFilter === null
              ? "bg-brand text-white"
              : "bg-muted text-muted-foreground hover:text-foreground"
          }`}
        >
          All
        </button>
        {categories.map((cat) => (
          <button
            key={cat}
            onClick={() => setCategoryFilter(cat)}
            className={`px-3 py-1 text-sm rounded-full transition-colors ${
              categoryFilter === cat
                ? "bg-brand text-white"
                : "bg-muted text-muted-foreground hover:text-foreground"
            }`}
          >
            {cat.charAt(0).toUpperCase() + cat.slice(1)}
          </button>
        ))}
      </div>

      {/* Results count */}
      <p className="text-sm text-muted-foreground mb-6">
        Showing {filteredMarkets.length} markets
      </p>

      {/* Markets Display */}
      {viewMode === "grid" ? (
        <MarketGrid markets={filteredMarkets} columns={3} variant="default" />
      ) : (
        <MarketList markets={filteredMarkets} showCategory={true} />
      )}
    </div>
  );
}

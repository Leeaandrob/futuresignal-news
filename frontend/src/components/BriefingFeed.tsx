import { useEffect, useState } from "react";
import { getTodayArticles, getTrendingMarkets, getBreakingMarkets, type Article, type Market } from "@/lib/api";
import { ArticleCard } from "@/components/ArticleCard";
import { MarketCard } from "@/components/MarketCard";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { formatChange, formatProbability } from "@/lib/utils";
import { Loader2, TrendingUp, TrendingDown, Minus } from "lucide-react";

interface BriefingData {
  articles: Article[];
  rising: Market[];
  falling: Market[];
  stable: Market[];
}

export function BriefingFeed() {
  const [data, setData] = useState<BriefingData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadBriefing() {
      try {
        const [articles, trending, breaking] = await Promise.all([
          getTodayArticles(),
          getTrendingMarkets(30),
          getBreakingMarkets(20),
        ]);

        // Combine and categorize markets
        const allMarkets = [...trending, ...breaking];
        const uniqueMarkets = allMarkets.filter(
          (m, i, arr) => arr.findIndex((x) => x.id === m.id) === i
        );

        // Categorize by change direction
        const rising = uniqueMarkets
          .filter((m) => m.change24h >= 0.02)
          .sort((a, b) => b.change24h - a.change24h)
          .slice(0, 5);

        const falling = uniqueMarkets
          .filter((m) => m.change24h <= -0.02)
          .sort((a, b) => a.change24h - b.change24h)
          .slice(0, 5);

        const stable = uniqueMarkets
          .filter((m) => Math.abs(m.change24h) < 0.02)
          .sort((a, b) => b.volume24h - a.volume24h)
          .slice(0, 3);

        setData({ articles, rising, falling, stable });
      } catch (err) {
        setError("Failed to load briefing. Please try again.");
        console.error(err);
      } finally {
        setLoading(false);
      }
    }

    loadBriefing();
  }, []);

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

  if (!data) return null;

  return (
    <div className="space-y-10">
      {/* Today's Stories */}
      {data.articles.length > 0 && (
        <section>
          <h2 className="font-semibold text-sm uppercase tracking-wider text-muted-foreground mb-4">
            Today's Stories
          </h2>
          <div className="space-y-4">
            {data.articles.slice(0, 5).map((article) => (
              <ArticleCard key={article.id} article={article} variant="compact" />
            ))}
          </div>
        </section>
      )}

      {/* Rising Confidence */}
      {data.rising.length > 0 && (
        <section>
          <h2 className="font-semibold text-sm uppercase tracking-wider text-muted-foreground mb-4 flex items-center gap-2">
            <TrendingUp className="w-4 h-4 text-bullish" />
            Rising Confidence
          </h2>
          <div className="space-y-3">
            {data.rising.map((market) => (
              <MarketRow key={market.id} market={market} direction="up" />
            ))}
          </div>
        </section>
      )}

      {/* Falling Confidence */}
      {data.falling.length > 0 && (
        <section>
          <h2 className="font-semibold text-sm uppercase tracking-wider text-muted-foreground mb-4 flex items-center gap-2">
            <TrendingDown className="w-4 h-4 text-bearish" />
            Falling Confidence
          </h2>
          <div className="space-y-3">
            {data.falling.map((market) => (
              <MarketRow key={market.id} market={market} direction="down" />
            ))}
          </div>
        </section>
      )}

      {/* Holding Steady */}
      {data.stable.length > 0 && (
        <section>
          <h2 className="font-semibold text-sm uppercase tracking-wider text-muted-foreground mb-4 flex items-center gap-2">
            <Minus className="w-4 h-4" />
            Holding Steady
          </h2>
          <div className="space-y-3">
            {data.stable.map((market) => (
              <MarketRow key={market.id} market={market} direction="stable" />
            ))}
          </div>
        </section>
      )}

      {/* No content state */}
      {data.articles.length === 0 && data.rising.length === 0 && data.falling.length === 0 && (
        <div className="text-center py-12">
          <p className="text-muted-foreground">
            No briefing content available yet. Check back later!
          </p>
        </div>
      )}
    </div>
  );
}

// Market row component for briefing
interface MarketRowProps {
  market: Market;
  direction: "up" | "down" | "stable";
}

function MarketRow({ market, direction }: MarketRowProps) {
  const category = market.category || "world";
  const change = market.change24h ?? 0;

  return (
    <a
      href={`/market/${market.slug}`}
      className="flex items-center justify-between p-3 border rounded-lg hover:bg-muted/50 transition-colors"
    >
      <div className="flex items-center gap-3 flex-1 min-w-0">
        <Badge variant={category as any} className="shrink-0">
          {category.toUpperCase()}
        </Badge>
        <span className="font-medium truncate">{market.question}</span>
      </div>
      <div className="flex items-center gap-3 shrink-0">
        <span className="font-mono font-bold">
          {formatProbability(market.probability)}
        </span>
        {direction === "up" && (
          <span className="font-mono text-bullish">
            {formatChange(change)}
          </span>
        )}
        {direction === "down" && (
          <span className="font-mono text-bearish">
            {formatChange(change)}
          </span>
        )}
        {direction === "stable" && (
          <span className="text-sm text-muted-foreground">(stable)</span>
        )}
      </div>
    </a>
  );
}

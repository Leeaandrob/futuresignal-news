import { useEffect, useState } from "react";
import {
  getArticlesByCategory,
  getMarketsByCategory,
  getTrendingMarkets,
  getBreakingMarkets,
  getNewMarkets,
  getTrendingArticles,
  getBreakingArticles,
  type Article,
  type Market,
} from "@/lib/api";
import { ArticleCard, ArticleGrid } from "@/components/ArticleCard";
import { MarketCard, MarketGrid, MarketList } from "@/components/MarketCard";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Loader2 } from "lucide-react";

interface CategoryFeedProps {
  category: string;
  isDynamic?: boolean;
}

export function CategoryFeed({ category, isDynamic = false }: CategoryFeedProps) {
  const [articles, setArticles] = useState<Article[]>([]);
  const [markets, setMarkets] = useState<Market[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadData() {
      setLoading(true);
      setError(null);

      try {
        let articlesData: Article[] = [];
        let marketsData: Market[] = [];

        // For dynamic categories, use specific endpoints
        if (isDynamic) {
          switch (category) {
            case "trending":
              [articlesData, marketsData] = await Promise.all([
                getTrendingArticles(20),
                getTrendingMarkets(30),
              ]);
              break;
            case "breaking":
              [articlesData, marketsData] = await Promise.all([
                getBreakingArticles(20),
                getBreakingMarkets(30),
              ]);
              break;
            case "new":
              marketsData = await getNewMarkets(30);
              break;
            default:
              [articlesData, marketsData] = await Promise.all([
                getArticlesByCategory(category, 20),
                getMarketsByCategory(category, 30),
              ]);
          }
        } else {
          // For static categories
          [articlesData, marketsData] = await Promise.all([
            getArticlesByCategory(category, 20),
            getMarketsByCategory(category, 30),
          ]);
        }

        setArticles(articlesData);
        setMarkets(marketsData);
      } catch (err) {
        setError("Failed to load content. Please try again.");
        console.error(err);
      } finally {
        setLoading(false);
      }
    }

    loadData();
  }, [category, isDynamic]);

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

  const hasArticles = articles.length > 0;
  const hasMarkets = markets.length > 0;

  // For "new" category, show markets only
  if (category === "new") {
    return (
      <div>
        {hasMarkets ? (
          <MarketGrid markets={markets} columns={3} variant="detailed" />
        ) : (
          <EmptyState category={category} />
        )}
      </div>
    );
  }

  return (
    <div className="grid gap-8 lg:grid-cols-3">
      {/* Main Content - Articles */}
      <div className="lg:col-span-2">
        {hasArticles ? (
          <section>
            <h2 className="headline-md mb-6">Latest Stories</h2>
            <ArticleGrid articles={articles} columns={2} />
          </section>
        ) : (
          <EmptyState category={category} type="articles" />
        )}
      </div>

      {/* Sidebar - Markets */}
      <aside>
        <Card className="sticky top-24">
          <CardHeader className="pb-3">
            <h2 className="font-semibold text-lg">
              {isDynamic ? "Live Markets" : "Top Markets"}
            </h2>
          </CardHeader>
          <CardContent className="p-0">
            {hasMarkets ? (
              <MarketList markets={markets.slice(0, 10)} showCategory={false} />
            ) : (
              <div className="p-4 text-center text-sm text-muted-foreground">
                No markets found
              </div>
            )}
          </CardContent>
        </Card>

        {/* View all markets link */}
        <div className="mt-4">
          <a
            href="/markets"
            className="block text-center py-3 border rounded-lg text-sm text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
          >
            Browse all markets →
          </a>
        </div>
      </aside>
    </div>
  );
}

// Empty state component
interface EmptyStateProps {
  category: string;
  type?: "articles" | "markets" | "all";
}

function EmptyState({ category, type = "all" }: EmptyStateProps) {
  const messages = {
    articles: `No stories yet in ${category}.`,
    markets: `No markets found in ${category}.`,
    all: `No content yet in ${category}.`,
  };

  return (
    <div className="text-center py-12 bg-muted/30 rounded-lg">
      <p className="text-muted-foreground mb-4">{messages[type]}</p>
      <p className="text-sm text-muted-foreground">
        Check back soon for the latest updates.
      </p>
    </div>
  );
}

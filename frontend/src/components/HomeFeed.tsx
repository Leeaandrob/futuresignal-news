import { useEffect, useState } from "react";
import { getHomeFeed, getSentiment, type HomeFeedResponse, type CategorySentiment } from "@/lib/api";
import { ArticleCard, ArticleGrid, ArticleList } from "@/components/ArticleCard";
import { MarketCard, MarketList, TrendingMarketsWidget } from "@/components/MarketCard";
import { MarketPulse } from "@/components/MarketPulse";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { formatFullDate } from "@/lib/utils";
import { Loader2, TrendingUp, Zap, Clock, ChevronRight } from "lucide-react";

export function HomeFeed() {
  const [feed, setFeed] = useState<HomeFeedResponse | null>(null);
  const [sentiments, setSentiments] = useState<CategorySentiment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadFeed() {
      try {
        const [feedData, sentimentData] = await Promise.all([
          getHomeFeed(),
          getSentiment(),
        ]);
        setFeed(feedData);
        setSentiments(sentimentData);
      } catch (err) {
        setError("Failed to load feed. Please try again.");
        console.error(err);
      } finally {
        setLoading(false);
      }
    }
    loadFeed();
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

  if (!feed) return null;

  const heroArticle = feed.featured?.[0];
  const secondaryFeatured = feed.featured?.slice(1, 3) || [];
  const today = formatFullDate(new Date());

  return (
    <div className="space-y-12">
      {/* Hero Section */}
      {heroArticle && (
        <section>
          <ArticleCard article={heroArticle} variant="hero" />
        </section>
      )}

      {/* Market Pulse - Category Momentum */}
      {sentiments.length > 0 && (
        <section>
          <MarketPulse sentiments={sentiments} />
        </section>
      )}

      {/* Featured Stories Grid */}
      {secondaryFeatured.length > 0 && (
        <section>
          <div className="grid gap-6 md:grid-cols-2">
            {secondaryFeatured.map((article) => (
              <ArticleCard key={article.id} article={article} variant="featured" />
            ))}
          </div>
        </section>
      )}

      {/* Main Content Grid */}
      <div className="grid gap-8 lg:grid-cols-3">
        {/* Left Column - Articles */}
        <div className="lg:col-span-2 space-y-8">
          {/* Today's Briefing */}
          {feed.today && feed.today.length > 0 && (
            <section>
              <Card className="border-brand/20 bg-gradient-to-r from-brand/5 to-transparent">
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Clock className="w-5 h-5 text-brand" />
                      <h2 className="font-bold text-lg">Today's Briefing</h2>
                    </div>
                    <span className="text-sm text-muted-foreground">{today}</span>
                  </div>
                </CardHeader>
                <CardContent className="p-0">
                  <ArticleList articles={feed.today.slice(0, 5)} />
                </CardContent>
              </Card>
            </section>
          )}

          {/* Recent Articles */}
          {feed.recent && feed.recent.length > 0 && (
            <section>
              <div className="flex items-center justify-between mb-6">
                <h2 className="headline-md">Recent Stories</h2>
                <a
                  href="/articles"
                  className="text-sm text-brand hover:underline flex items-center gap-1"
                >
                  View all <ChevronRight className="w-4 h-4" />
                </a>
              </div>
              <ArticleGrid articles={feed.recent.slice(0, 6)} columns={2} />
            </section>
          )}
        </div>

        {/* Right Column - Sidebar */}
        <aside className="space-y-6">
          {/* Trending Markets */}
          {feed.trendingMarkets && feed.trendingMarkets.length > 0 && (
            <section>
              <Card>
                <CardHeader className="pb-2">
                  <div className="flex items-center gap-2">
                    <TrendingUp className="w-5 h-5 text-trending" />
                    <h3 className="font-semibold text-lg">Trending Markets</h3>
                  </div>
                </CardHeader>
                <CardContent className="p-0">
                  <MarketList
                    markets={feed.trendingMarkets.slice(0, 5)}
                    showCategory={false}
                  />
                </CardContent>
              </Card>
            </section>
          )}

          {/* Breaking Now */}
          <section>
            <Card className="border-breaking/20">
              <CardHeader className="pb-2">
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 bg-breaking rounded-full animate-pulse" />
                  <Zap className="w-5 h-5 text-breaking" />
                  <h3 className="font-semibold text-lg">Breaking</h3>
                </div>
              </CardHeader>
              <CardContent>
                <a
                  href="/category/breaking"
                  className="block text-sm text-muted-foreground hover:text-foreground transition-colors"
                >
                  View all breaking news →
                </a>
              </CardContent>
            </Card>
          </section>

          {/* Categories Quick Links */}
          <section>
            <Card>
              <CardHeader className="pb-3">
                <h3 className="font-semibold text-lg">Browse Categories</h3>
              </CardHeader>
              <CardContent className="flex flex-wrap gap-2">
                <a href="/category/politics">
                  <Badge variant="politics">Politics</Badge>
                </a>
                <a href="/category/crypto">
                  <Badge variant="crypto">Crypto</Badge>
                </a>
                <a href="/category/tech">
                  <Badge variant="tech">Tech</Badge>
                </a>
                <a href="/category/sports">
                  <Badge variant="sports">Sports</Badge>
                </a>
                <a href="/category/finance">
                  <Badge variant="finance">Finance</Badge>
                </a>
                <a href="/category/world">
                  <Badge variant="world">World</Badge>
                </a>
              </CardContent>
            </Card>
          </section>
        </aside>
      </div>
    </div>
  );
}

// Hero article component for SSR
interface HeroArticleProps {
  article: HomeFeedResponse["featured"][0];
}

export function HeroArticle({ article }: HeroArticleProps) {
  return <ArticleCard article={article} variant="hero" />;
}

// Empty state for when there's no content
export function EmptyFeed() {
  return (
    <div className="text-center py-20">
      <h2 className="text-2xl font-bold mb-4">No content yet</h2>
      <p className="text-muted-foreground mb-6">
        We're generating content from prediction markets. Check back soon!
      </p>
      <a href="/markets" className="text-brand hover:underline">
        Browse all markets →
      </a>
    </div>
  );
}

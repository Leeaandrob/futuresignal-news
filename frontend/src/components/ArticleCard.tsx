import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn, formatTimeAgo, formatVolume, getArticleUrl } from "@/lib/utils";
import { getArticleTypeBadge, type Article, type ArticleType } from "@/lib/types";
import { TrendingUp, TrendingDown, Zap, Clock, Eye } from "lucide-react";

interface ArticleCardProps {
  article: Article;
  variant?: "default" | "featured" | "compact" | "hero";
}

export function ArticleCard({ article, variant = "default" }: ArticleCardProps) {
  const typeBadge = getArticleTypeBadge(article.type);
  const hasChange = article.markets?.[0]?.change24h !== undefined && article.markets?.[0]?.change24h !== null;
  const change = article.markets?.[0]?.change24h ?? 0;
  const isPositive = change >= 0;
  const category = article.category || "world";

  // Hero variant - large featured card
  if (variant === "hero") {
    return (
      <a href={getArticleUrl(article.slug)}>
        <Card className="border-2 border-brand/20 bg-gradient-to-br from-white to-brand/5 hover:shadow-lg transition-all">
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                {article.type === "breaking" && (
                  <span className="flex items-center gap-1 text-breaking text-sm font-semibold animate-pulse">
                    <Zap className="w-4 h-4" />
                    BREAKING
                  </span>
                )}
                <Badge variant={typeBadge.variant as any}>
                  {typeBadge.label}
                </Badge>
                <Badge variant={category as any}>
                  {category.toUpperCase()}
                </Badge>
              </div>
              {hasChange && (
                <div
                  className={cn(
                    "flex items-center gap-1 text-2xl font-mono font-bold",
                    isPositive ? "text-bullish" : "text-bearish"
                  )}
                >
                  {isPositive ? (
                    <TrendingUp className="w-6 h-6" />
                  ) : (
                    <TrendingDown className="w-6 h-6" />
                  )}
                  {isPositive ? "+" : ""}
                  {(change * 100).toFixed(1)}%
                </div>
              )}
            </div>
          </CardHeader>
          <CardContent>
            <h2 className="headline-lg mb-3">{article.title}</h2>
            {article.subtitle && (
              <p className="text-lg text-muted-foreground mb-4">
                {article.subtitle}
              </p>
            )}
            <p className="text-muted-foreground body-md mb-4 line-clamp-2">
              {article.summary}
            </p>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <span className="flex items-center gap-1">
                <Clock className="w-4 h-4" />
                {formatTimeAgo(article.publishedAt)}
              </span>
              {article.views > 0 && (
                <span className="flex items-center gap-1">
                  <Eye className="w-4 h-4" />
                  {article.views} views
                </span>
              )}
              {article.markets?.[0]?.volume24h && (
                <span>{formatVolume(article.markets[0].volume24h)} volume</span>
              )}
            </div>
          </CardContent>
        </Card>
      </a>
    );
  }

  // Featured variant - medium size with more detail
  if (variant === "featured") {
    return (
      <a href={getArticleUrl(article.slug)}>
        <Card className="h-full hover:shadow-md hover:border-brand/30 transition-all">
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Badge variant={typeBadge.variant as any}>
                  {typeBadge.label}
                </Badge>
                <Badge variant={category as any}>
                  {category.toUpperCase()}
                </Badge>
              </div>
              {hasChange && (
                <div
                  className={cn(
                    "flex items-center gap-1 font-mono font-bold text-lg",
                    isPositive ? "text-bullish" : "text-bearish"
                  )}
                >
                  {isPositive ? "+" : ""}
                  {(change * 100).toFixed(1)}%
                </div>
              )}
            </div>
          </CardHeader>
          <CardContent>
            <h3 className="font-semibold text-xl leading-tight mb-2 line-clamp-2">
              {article.title}
            </h3>
            <p className="text-muted-foreground text-sm mb-4 line-clamp-2">
              {article.summary}
            </p>
            <div className="flex items-center gap-3 text-sm text-muted-foreground">
              <span>{formatTimeAgo(article.publishedAt)}</span>
              {article.markets?.length > 0 && (
                <span>{article.markets.length} market{article.markets.length > 1 ? "s" : ""}</span>
              )}
            </div>
          </CardContent>
        </Card>
      </a>
    );
  }

  // Compact variant - for lists
  if (variant === "compact") {
    return (
      <a
        href={getArticleUrl(article.slug)}
        className="flex items-start gap-2 p-3 border-b hover:bg-muted/50 transition-colors overflow-hidden"
      >
        <Badge variant={category as any} className="shrink-0 mt-0.5">
          {category.toUpperCase().slice(0, 4)}
        </Badge>
        <span className="text-sm flex-1 line-clamp-2 break-words">{article.title}</span>
        <span className="text-xs text-muted-foreground shrink-0 mt-0.5">
          {formatTimeAgo(article.publishedAt)}
        </span>
      </a>
    );
  }

  // Default card variant
  return (
    <a href={getArticleUrl(article.slug)}>
      <Card className="h-full hover:shadow-md hover:border-brand/30 transition-all cursor-pointer">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Badge variant={typeBadge.variant as any} className="text-xs">
                {typeBadge.label}
              </Badge>
              <Badge variant={article.category as any} className="text-xs">
                {article.category.toUpperCase()}
              </Badge>
            </div>
            {hasChange && (
              <div
                className={cn(
                  "flex items-center gap-1 font-mono font-bold",
                  isPositive ? "text-bullish" : "text-bearish"
                )}
              >
                {isPositive ? "+" : ""}
                {(change * 100).toFixed(1)}%
              </div>
            )}
          </div>
        </CardHeader>
        <CardContent>
          <h3 className="font-semibold text-lg leading-tight mb-2 line-clamp-2">
            {article.title}
          </h3>
          <p className="text-muted-foreground text-sm mb-3 line-clamp-2">
            {article.summary}
          </p>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <span>{formatTimeAgo(article.publishedAt)}</span>
          </div>
        </CardContent>
      </Card>
    </a>
  );
}

// Grid component for article cards
interface ArticleGridProps {
  articles: Article[];
  columns?: 2 | 3 | 4;
  variant?: "default" | "featured";
}

export function ArticleGrid({
  articles,
  columns = 3,
  variant = "default",
}: ArticleGridProps) {
  const colsClass = {
    2: "md:grid-cols-2",
    3: "md:grid-cols-2 lg:grid-cols-3",
    4: "md:grid-cols-2 lg:grid-cols-4",
  };

  return (
    <div className={cn("grid gap-6", colsClass[columns])}>
      {articles.map((article) => (
        <ArticleCard key={article.id} article={article} variant={variant} />
      ))}
    </div>
  );
}

// List component for compact articles
interface ArticleListProps {
  articles: Article[];
}

export function ArticleList({ articles }: ArticleListProps) {
  return (
    <div className="border rounded-lg divide-y">
      {articles.map((article) => (
        <ArticleCard key={article.id} article={article} variant="compact" />
      ))}
    </div>
  );
}

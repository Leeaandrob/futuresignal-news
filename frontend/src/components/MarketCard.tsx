import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ProbabilityBar, ProbabilityInline, OutcomePrices } from "@/components/ProbabilityBar";
import {
  cn,
  formatTimeAgo,
  formatVolume,
  getMarketUrl,
  getSignificance,
} from "@/lib/utils";
import type { Market } from "@/lib/types";
import { TrendingUp, TrendingDown, ExternalLink, Clock, DollarSign, MessageCircle, Users, BarChart3 } from "lucide-react";

// Market image component with fallback
function MarketImage({ src, alt, size = "md" }: { src?: string; alt: string; size?: "sm" | "md" | "lg" }) {
  const sizeClasses = {
    sm: "w-8 h-8",
    md: "w-12 h-12",
    lg: "w-16 h-16",
  };

  if (!src) {
    return (
      <div className={cn(sizeClasses[size], "rounded-lg bg-muted flex items-center justify-center shrink-0")}>
        <BarChart3 className="w-1/2 h-1/2 text-muted-foreground" />
      </div>
    );
  }

  return (
    <img
      src={src}
      alt={alt}
      className={cn(sizeClasses[size], "rounded-lg object-cover shrink-0")}
      onError={(e) => {
        e.currentTarget.style.display = "none";
      }}
    />
  );
}

interface MarketCardProps {
  market: Market;
  variant?: "default" | "compact" | "detailed" | "mini";
  showCategory?: boolean;
}

export function MarketCard({
  market,
  variant = "default",
  showCategory = true,
}: MarketCardProps) {
  const change = market.change24h ?? 0;
  const probability = market.probability ?? 0;
  const category = market.category || "world";
  const isPositive = change >= 0;
  const significance = getSignificance(change);

  // Mini variant - just probability and change
  if (variant === "mini") {
    return (
      <div className="flex items-center justify-between p-2 rounded-lg bg-muted/50">
        <span className="text-sm truncate max-w-[60%]">{market.question}</span>
        <ProbabilityInline
          probability={probability}
          change={change}
          size="sm"
        />
      </div>
    );
  }

  // Compact variant - for lists and sidebars
  if (variant === "compact") {
    return (
      <a
        href={getMarketUrl(market.slug)}
        className="flex items-center gap-3 p-3 border-b hover:bg-muted/50 transition-colors overflow-hidden"
      >
        <MarketImage src={market.image} alt={market.question} size="sm" />
        <div className="flex-1 min-w-0">
          <span className="text-sm line-clamp-1 break-words">{market.question}</span>
          <div className="flex items-center gap-2 mt-0.5">
            <span className="font-mono font-bold text-sm">
              {(probability * 100).toFixed(0)}%
            </span>
            <span
              className={cn(
                "font-mono text-xs",
                isPositive ? "text-bullish" : "text-bearish"
              )}
            >
              {isPositive ? "+" : ""}{(change * 100).toFixed(1)}%
            </span>
          </div>
        </div>
        <span className="text-xs text-muted-foreground shrink-0">
          {formatVolume(market.volume24h)}
        </span>
      </a>
    );
  }

  // Detailed variant - full info with probability bar
  if (variant === "detailed") {
    const change7d = market.change7d ?? 0;
    const isPositive7d = change7d >= 0;

    return (
      <Card className="h-full hover:shadow-md transition-all">
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              {showCategory && (
                <Badge variant={category as any}>
                  {category.toUpperCase()}
                </Badge>
              )}
              {significance === "breaking" && (
                <Badge variant="breaking">MOVING</Badge>
              )}
            </div>
            <a
              href={market.polymarketUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-muted-foreground hover:text-foreground transition-colors"
              onClick={(e) => e.stopPropagation()}
            >
              <ExternalLink className="w-4 h-4" />
            </a>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <a href={getMarketUrl(market.slug)} className="block group">
            <div className="flex gap-3">
              <MarketImage src={market.image} alt={market.question} size="lg" />
              <div className="flex-1 min-w-0">
                <h3 className="font-semibold text-lg leading-tight group-hover:text-brand transition-colors line-clamp-2">
                  {market.question}
                </h3>
                {market.eventTitle && (
                  <p className="text-sm text-muted-foreground mt-1 truncate">
                    {market.eventTitle}
                  </p>
                )}
              </div>
            </div>
          </a>

          <ProbabilityBar
            probability={probability}
            previousProb={market.previousProb}
            showChange={true}
          />

          {market.outcomes && market.outcomes.length > 2 && (
            <OutcomePrices
              outcomes={market.outcomes}
              prices={market.outcomePrices || []}
              size="sm"
            />
          )}

          {/* Volume and Change Stats */}
          <div className="grid grid-cols-2 gap-2 text-sm">
            <div className="flex items-center justify-between p-2 bg-muted/50 rounded">
              <span className="text-muted-foreground">24h</span>
              <div className="flex items-center gap-2">
                <span className={cn("font-mono font-medium", isPositive ? "text-bullish" : "text-bearish")}>
                  {isPositive ? "+" : ""}{(change * 100).toFixed(1)}%
                </span>
                <span className="text-muted-foreground">{formatVolume(market.volume24h)}</span>
              </div>
            </div>
            {market.volume7d !== undefined && (
              <div className="flex items-center justify-between p-2 bg-muted/50 rounded">
                <span className="text-muted-foreground">7d</span>
                <div className="flex items-center gap-2">
                  <span className={cn("font-mono font-medium", isPositive7d ? "text-bullish" : "text-bearish")}>
                    {isPositive7d ? "+" : ""}{(change7d * 100).toFixed(1)}%
                  </span>
                  <span className="text-muted-foreground">{formatVolume(market.volume7d)}</span>
                </div>
              </div>
            )}
          </div>

          {/* Engagement Metrics */}
          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            {market.commentCount !== undefined && market.commentCount > 0 && (
              <span className="flex items-center gap-1">
                <MessageCircle className="w-4 h-4" />
                {market.commentCount}
              </span>
            )}
            {market.competitorCount !== undefined && market.competitorCount > 0 && (
              <span className="flex items-center gap-1">
                <Users className="w-4 h-4" />
                {market.competitorCount.toLocaleString()}
              </span>
            )}
            <span className="flex items-center gap-1 ml-auto">
              <Clock className="w-4 h-4" />
              {formatTimeAgo(market.updatedAt)}
            </span>
          </div>

          {/* Polymarket Tags */}
          {market.polymarketTags && market.polymarketTags.length > 0 && (
            <div className="flex flex-wrap gap-1">
              {market.polymarketTags.slice(0, 3).map((tag) => (
                <Badge key={tag.slug} variant="secondary" className="text-xs">
                  {tag.label}
                </Badge>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    );
  }

  // Default card variant
  return (
    <a href={getMarketUrl(market.slug)}>
      <Card className="h-full hover:shadow-md hover:border-brand/30 transition-all cursor-pointer">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              {showCategory && (
                <Badge variant={category as any} className="text-xs">
                  {category.toUpperCase()}
                </Badge>
              )}
            </div>
            <div
              className={cn(
                "flex items-center gap-1 font-mono font-bold",
                isPositive ? "text-bullish" : "text-bearish"
              )}
            >
              {isPositive ? <TrendingUp className="w-4 h-4" /> : <TrendingDown className="w-4 h-4" />}
              {isPositive ? "+" : ""}
              {(change * 100).toFixed(1)}%
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex gap-3 mb-3">
            <MarketImage src={market.image} alt={market.question} size="md" />
            <h3 className="font-semibold leading-tight line-clamp-2 flex-1">
              {market.question}
            </h3>
          </div>
          <div className="flex items-center justify-between">
            <span className="font-mono font-bold text-xl">
              {(probability * 100).toFixed(0)}%
            </span>
            <div className="flex items-center gap-3 text-sm text-muted-foreground">
              {market.commentCount !== undefined && market.commentCount > 0 && (
                <span className="flex items-center gap-1">
                  <MessageCircle className="w-3 h-3" />
                  {market.commentCount}
                </span>
              )}
              <span>{formatVolume(market.volume24h)}</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </a>
  );
}

// Grid component for market cards
interface MarketGridProps {
  markets: Market[];
  columns?: 2 | 3 | 4;
  variant?: "default" | "detailed";
  showCategory?: boolean;
}

export function MarketGrid({
  markets,
  columns = 3,
  variant = "default",
  showCategory = true,
}: MarketGridProps) {
  const colsClass = {
    2: "md:grid-cols-2",
    3: "md:grid-cols-2 lg:grid-cols-3",
    4: "md:grid-cols-2 lg:grid-cols-4",
  };

  return (
    <div className={cn("grid gap-6", colsClass[columns])}>
      {markets.map((market) => (
        <MarketCard
          key={market.id}
          market={market}
          variant={variant}
          showCategory={showCategory}
        />
      ))}
    </div>
  );
}

// List component for compact markets
interface MarketListProps {
  markets: Market[];
  showCategory?: boolean;
}

export function MarketList({ markets, showCategory = true }: MarketListProps) {
  return (
    <div className="border rounded-lg divide-y">
      {markets.map((market) => (
        <MarketCard
          key={market.id}
          market={market}
          variant="compact"
          showCategory={showCategory}
        />
      ))}
    </div>
  );
}

// Trending markets sidebar widget
interface TrendingMarketsWidgetProps {
  markets: Market[];
  title?: string;
}

export function TrendingMarketsWidget({
  markets,
  title = "Trending Markets",
}: TrendingMarketsWidgetProps) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <h3 className="font-semibold text-lg">{title}</h3>
      </CardHeader>
      <CardContent className="p-0">
        <div className="divide-y">
          {markets.slice(0, 5).map((market) => (
            <MarketCard
              key={market.id}
              market={market}
              variant="compact"
              showCategory={false}
            />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

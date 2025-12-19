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
import { TrendingUp, TrendingDown, ExternalLink, Clock, DollarSign } from "lucide-react";

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
        className="flex items-start gap-2 p-3 border-b hover:bg-muted/50 transition-colors overflow-hidden"
      >
        <div
          className={cn(
            "flex items-center gap-1 font-mono font-semibold text-xs shrink-0 mt-0.5",
            isPositive ? "text-bullish" : "text-bearish"
          )}
        >
          {isPositive ? <TrendingUp className="w-3 h-3" /> : <TrendingDown className="w-3 h-3" />}
          {isPositive ? "+" : ""}
          {(change * 100).toFixed(1)}%
        </div>
        <span className="font-mono font-bold text-sm shrink-0 mt-0.5">
          {(probability * 100).toFixed(0)}%
        </span>
        <span className="text-sm flex-1 line-clamp-2 break-words">{market.question}</span>
        <span className="text-xs text-muted-foreground shrink-0 mt-0.5">
          {formatVolume(market.volume24h)}
        </span>
      </a>
    );
  }

  // Detailed variant - full info with probability bar
  if (variant === "detailed") {
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
            <h3 className="font-semibold text-lg leading-tight group-hover:text-brand transition-colors line-clamp-2">
              {market.question}
            </h3>
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

          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            <span className="flex items-center gap-1">
              <DollarSign className="w-4 h-4" />
              {formatVolume(market.volume24h)} 24h
            </span>
            <span className="flex items-center gap-1">
              <Clock className="w-4 h-4" />
              {formatTimeAgo(market.updatedAt)}
            </span>
          </div>
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
          <h3 className="font-semibold leading-tight mb-3 line-clamp-2">
            {market.question}
          </h3>
          <div className="flex items-center justify-between">
            <span className="font-mono font-bold text-xl">
              {(probability * 100).toFixed(0)}%
            </span>
            <span className="text-sm text-muted-foreground">
              {formatVolume(market.volume24h)}
            </span>
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

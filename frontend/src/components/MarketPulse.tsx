import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn, formatVolume, getMarketUrl } from "@/lib/utils";
import type { CategorySentiment } from "@/lib/types";
import { TrendingUp, TrendingDown, Activity, Zap } from "lucide-react";

// =============================================================================
// MARKET PULSE DASHBOARD (Homepage)
// =============================================================================

interface MarketPulseProps {
  sentiments: CategorySentiment[];
  title?: string;
  showTopMover?: boolean;
}

export function MarketPulse({
  sentiments,
  title = "Market Pulse",
  showTopMover = true,
}: MarketPulseProps) {
  // Sort by absolute momentum for visual impact
  const sorted = [...sentiments].sort(
    (a, b) => Math.abs(b.momentum) - Math.abs(a.momentum)
  );

  // Calculate overall market sentiment
  const totalVolume = sentiments.reduce((sum, s) => sum + s.totalVolume24h, 0);
  const weightedMomentum =
    totalVolume > 0
      ? sentiments.reduce((sum, s) => sum + s.momentum * s.totalVolume24h, 0) /
        totalVolume
      : 0;
  const totalBreaking = sentiments.reduce((sum, s) => sum + s.breakingCount, 0);

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Activity className="w-5 h-5 text-brand" />
            <h2 className="font-bold text-lg">{title}</h2>
          </div>
          <div className="flex items-center gap-3 text-sm">
            {totalBreaking > 0 && (
              <Badge variant="breaking" className="gap-1">
                <Zap className="w-3 h-3" />
                {totalBreaking} Breaking
              </Badge>
            )}
            <span
              className={cn(
                "font-mono font-bold",
                weightedMomentum >= 0 ? "text-bullish" : "text-bearish"
              )}
            >
              {weightedMomentum >= 0 ? "+" : ""}
              {(weightedMomentum * 100).toFixed(2)}%
            </span>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-3 pt-0">
        {sorted.map((sentiment) => (
          <MomentumBar
            key={sentiment.category}
            sentiment={sentiment}
            showTopMover={showTopMover}
          />
        ))}
      </CardContent>
    </Card>
  );
}

// =============================================================================
// MOMENTUM BAR (Individual Category Row)
// =============================================================================

interface MomentumBarProps {
  sentiment: CategorySentiment;
  showTopMover?: boolean;
}

function MomentumBar({ sentiment, showTopMover = true }: MomentumBarProps) {
  const { momentum, breakingCount, topMover, topMoverSlug, topMoverChange } =
    sentiment;
  const isBullish = momentum >= 0;
  const absPercent = Math.min(Math.abs(momentum) * 100 * 5, 50); // Scale for visibility (max 50%)

  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-sm">
        <div className="flex items-center gap-2">
          <a
            href={`/category/${sentiment.category}`}
            className="font-medium hover:text-brand transition-colors"
          >
            {sentiment.name}
          </a>
          {breakingCount > 0 && (
            <Badge variant="secondary" className="text-xs py-0 h-5">
              {breakingCount} moving
            </Badge>
          )}
        </div>
        <div className="flex items-center gap-2">
          <span className="text-muted-foreground text-xs">
            {sentiment.marketCount} markets
          </span>
          <span className="text-muted-foreground text-xs">
            {formatVolume(sentiment.totalVolume24h)}
          </span>
          <span
            className={cn(
              "font-mono font-bold min-w-[60px] text-right",
              isBullish ? "text-bullish" : "text-bearish"
            )}
          >
            {isBullish ? (
              <TrendingUp className="w-3 h-3 inline mr-1" />
            ) : (
              <TrendingDown className="w-3 h-3 inline mr-1" />
            )}
            {isBullish ? "+" : ""}
            {(momentum * 100).toFixed(2)}%
          </span>
        </div>
      </div>

      {/* Momentum Bar Visualization */}
      <div className="relative h-2 bg-muted rounded-full overflow-hidden">
        <div
          className={cn(
            "absolute top-0 h-full rounded-full transition-all duration-500",
            isBullish ? "bg-bullish left-1/2" : "bg-bearish right-1/2"
          )}
          style={{
            width: `${absPercent}%`,
            ...(isBullish ? { left: "50%" } : { right: "50%" }),
          }}
        />
        {/* Center line */}
        <div className="absolute left-1/2 top-0 w-px h-full bg-border" />
      </div>

      {/* Top Mover */}
      {showTopMover && topMover && topMoverSlug && (
        <a
          href={getMarketUrl(topMoverSlug)}
          className="flex items-center gap-2 text-xs text-muted-foreground hover:text-foreground transition-colors pl-2"
        >
          <span className="truncate max-w-[250px]">{topMover}</span>
          <span
            className={cn(
              "font-mono shrink-0",
              topMoverChange >= 0 ? "text-bullish" : "text-bearish"
            )}
          >
            {topMoverChange >= 0 ? "+" : ""}
            {(topMoverChange * 100).toFixed(1)}%
          </span>
        </a>
      )}
    </div>
  );
}

// =============================================================================
// PULSE HEADER BAR (Ultra-Compact for Article Pages)
// =============================================================================

interface PulseHeaderBarProps {
  sentiments: CategorySentiment[];
  currentCategory?: string;
}

export function PulseHeaderBar({
  sentiments,
  currentCategory,
}: PulseHeaderBarProps) {
  // Show top 5 by volume, with current category first if present
  const sorted = [...sentiments]
    .sort((a, b) => b.totalVolume24h - a.totalVolume24h)
    .slice(0, 6);

  // Calculate overall momentum
  const totalVolume = sentiments.reduce((sum, s) => sum + s.totalVolume24h, 0);
  const weightedMomentum =
    totalVolume > 0
      ? sentiments.reduce((sum, s) => sum + s.momentum * s.totalVolume24h, 0) /
        totalVolume
      : 0;

  return (
    <div className="flex items-center gap-4 py-2 px-4 bg-muted/50 rounded-lg overflow-x-auto scrollbar-hide">
      <div className="flex items-center gap-2 shrink-0">
        <Activity className="w-4 h-4 text-brand" />
        <span className="text-xs font-medium text-muted-foreground">PULSE</span>
        <span
          className={cn(
            "font-mono text-sm font-bold",
            weightedMomentum >= 0 ? "text-bullish" : "text-bearish"
          )}
        >
          {weightedMomentum >= 0 ? "+" : ""}
          {(weightedMomentum * 100).toFixed(1)}%
        </span>
      </div>
      <div className="w-px h-4 bg-border shrink-0" />
      <div className="flex items-center gap-3">
        {sorted.map((s) => (
          <a
            key={s.category}
            href={`/category/${s.category}`}
            className={cn(
              "flex items-center gap-1.5 text-xs hover:opacity-80 transition-opacity shrink-0",
              currentCategory === s.category && "font-bold"
            )}
          >
            <span className="text-muted-foreground">{s.name}</span>
            <span
              className={cn(
                "font-mono font-medium",
                s.momentum >= 0 ? "text-bullish" : "text-bearish"
              )}
            >
              {s.momentum >= 0 ? "+" : ""}
              {(s.momentum * 100).toFixed(1)}%
            </span>
          </a>
        ))}
      </div>
    </div>
  );
}

// =============================================================================
// CATEGORY SENTIMENT CARD (Individual Category Detail)
// =============================================================================

interface CategorySentimentCardProps {
  sentiment: CategorySentiment;
  variant?: "default" | "compact";
}

export function CategorySentimentCard({
  sentiment,
  variant = "default",
}: CategorySentimentCardProps) {
  const { momentum, breakingCount, topMover, topMoverSlug, topMoverChange } =
    sentiment;
  const isBullish = momentum >= 0;
  const absPercent = Math.min(Math.abs(momentum) * 100 * 5, 50);

  if (variant === "compact") {
    return (
      <a
        href={`/category/${sentiment.category}`}
        className="flex items-center gap-3 p-3 border rounded-lg hover:bg-muted/50 transition-colors"
      >
        <div
          className="w-10 h-10 rounded-lg flex items-center justify-center"
          style={{ backgroundColor: `${sentiment.color}20` }}
        >
          {isBullish ? (
            <TrendingUp className="w-5 h-5" style={{ color: sentiment.color }} />
          ) : (
            <TrendingDown
              className="w-5 h-5"
              style={{ color: sentiment.color }}
            />
          )}
        </div>
        <div className="flex-1 min-w-0">
          <div className="font-medium">{sentiment.name}</div>
          <div className="text-xs text-muted-foreground">
            {sentiment.marketCount} markets
          </div>
        </div>
        <div className="text-right">
          <div
            className={cn(
              "font-mono font-bold",
              isBullish ? "text-bullish" : "text-bearish"
            )}
          >
            {isBullish ? "+" : ""}
            {(momentum * 100).toFixed(2)}%
          </div>
          <div className="text-xs text-muted-foreground">
            {formatVolume(sentiment.totalVolume24h)}
          </div>
        </div>
      </a>
    );
  }

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div
              className="w-8 h-8 rounded-lg flex items-center justify-center"
              style={{ backgroundColor: `${sentiment.color}20` }}
            >
              {isBullish ? (
                <TrendingUp
                  className="w-4 h-4"
                  style={{ color: sentiment.color }}
                />
              ) : (
                <TrendingDown
                  className="w-4 h-4"
                  style={{ color: sentiment.color }}
                />
              )}
            </div>
            <a
              href={`/category/${sentiment.category}`}
              className="font-bold hover:text-brand transition-colors"
            >
              {sentiment.name}
            </a>
          </div>
          <div className="flex items-center gap-2">
            {breakingCount > 0 && (
              <Badge variant="breaking" className="text-xs">
                {breakingCount} moving
              </Badge>
            )}
            <span
              className={cn(
                "font-mono font-bold text-lg",
                isBullish ? "text-bullish" : "text-bearish"
              )}
            >
              {isBullish ? "+" : ""}
              {(momentum * 100).toFixed(2)}%
            </span>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {/* Stats Row */}
        <div className="flex items-center gap-4 text-sm text-muted-foreground">
          <span>{sentiment.marketCount} markets</span>
          <span>{formatVolume(sentiment.totalVolume24h)} 24h vol</span>
        </div>

        {/* Momentum Bar */}
        <div className="relative h-3 bg-muted rounded-full overflow-hidden">
          <div
            className={cn(
              "absolute top-0 h-full rounded-full transition-all duration-500",
              isBullish ? "bg-bullish" : "bg-bearish"
            )}
            style={{
              width: `${absPercent}%`,
              ...(isBullish ? { left: "50%" } : { right: "50%" }),
            }}
          />
          <div className="absolute left-1/2 top-0 w-px h-full bg-border" />
        </div>

        {/* Top Mover */}
        {topMover && topMoverSlug && (
          <a
            href={getMarketUrl(topMoverSlug)}
            className="block p-2 bg-muted/50 rounded-lg hover:bg-muted transition-colors"
          >
            <div className="text-xs text-muted-foreground mb-1">Top Mover</div>
            <div className="flex items-center justify-between gap-2">
              <span className="text-sm truncate">{topMover}</span>
              <span
                className={cn(
                  "font-mono font-bold shrink-0",
                  topMoverChange >= 0 ? "text-bullish" : "text-bearish"
                )}
              >
                {topMoverChange >= 0 ? "+" : ""}
                {(topMoverChange * 100).toFixed(1)}%
              </span>
            </div>
          </a>
        )}
      </CardContent>
    </Card>
  );
}

// =============================================================================
// CATEGORY SENTIMENT GRID
// =============================================================================

interface CategorySentimentGridProps {
  sentiments: CategorySentiment[];
  columns?: 2 | 3 | 4;
  variant?: "default" | "compact";
}

export function CategorySentimentGrid({
  sentiments,
  columns = 3,
  variant = "default",
}: CategorySentimentGridProps) {
  const colsClass = {
    2: "md:grid-cols-2",
    3: "md:grid-cols-2 lg:grid-cols-3",
    4: "md:grid-cols-2 lg:grid-cols-4",
  };

  return (
    <div className={cn("grid gap-4", colsClass[columns])}>
      {sentiments.map((sentiment) => (
        <CategorySentimentCard
          key={sentiment.category}
          sentiment={sentiment}
          variant={variant}
        />
      ))}
    </div>
  );
}

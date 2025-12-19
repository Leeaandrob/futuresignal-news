import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn, formatTimeAgo } from "@/lib/utils";
import type { SocialSignal, MarketMovement } from "@/lib/types";
import { TrendingUp, TrendingDown, ExternalLink, CheckCircle } from "lucide-react";

interface SocialSignalCardProps {
  signal: SocialSignal;
  variant?: "default" | "compact" | "inline";
}

export function SocialSignalCard({ signal, variant = "default" }: SocialSignalCardProps) {
  const isPositiveImpact = signal.market_impact >= 0;
  const impactPercent = Math.abs(signal.market_impact * 100);

  // Compact inline variant for article headers
  if (variant === "inline") {
    return (
      <a
        href={signal.tweet_url}
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex items-center gap-2 px-3 py-1.5 bg-muted/50 rounded-full hover:bg-muted transition-colors text-sm"
      >
        <img
          src={signal.avatar_url || `https://unavatar.io/twitter/${signal.handle}`}
          alt={signal.name}
          className="w-5 h-5 rounded-full"
        />
        <span className="font-medium">@{signal.handle}</span>
        {signal.verified && (
          <CheckCircle className="w-3.5 h-3.5 text-blue-500" />
        )}
        <ExternalLink className="w-3 h-3 text-muted-foreground" />
      </a>
    );
  }

  // Compact variant for sidebars/lists
  if (variant === "compact") {
    return (
      <a
        href={signal.tweet_url}
        target="_blank"
        rel="noopener noreferrer"
        className="flex items-start gap-3 p-3 border-b hover:bg-muted/50 transition-colors"
      >
        <img
          src={signal.avatar_url || `https://unavatar.io/twitter/${signal.handle}`}
          alt={signal.name}
          className="w-10 h-10 rounded-full shrink-0"
        />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-1.5 mb-1">
            <span className="font-semibold text-sm">{signal.name}</span>
            {signal.verified && (
              <CheckCircle className="w-3.5 h-3.5 text-blue-500" />
            )}
            <span className="text-xs text-muted-foreground">@{signal.handle}</span>
          </div>
          <p className="text-sm text-muted-foreground line-clamp-2">{signal.content}</p>
          <div className="flex items-center gap-2 mt-1 text-xs text-muted-foreground">
            <span>{formatTimeAgo(signal.posted_at)}</span>
            {impactPercent >= 1 && (
              <span className={cn(
                "flex items-center gap-0.5 font-mono font-medium",
                isPositiveImpact ? "text-bullish" : "text-bearish"
              )}>
                {isPositiveImpact ? <TrendingUp className="w-3 h-3" /> : <TrendingDown className="w-3 h-3" />}
                {impactPercent.toFixed(1)}%
              </span>
            )}
          </div>
        </div>
      </a>
    );
  }

  // Default full card variant
  return (
    <Card className="hover:shadow-md hover:border-brand/30 transition-all">
      <CardHeader className="pb-3">
        <div className="flex items-start gap-3">
          <img
            src={signal.avatar_url || `https://unavatar.io/twitter/${signal.handle}`}
            alt={signal.name}
            className="w-12 h-12 rounded-full"
          />
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <span className="font-bold">{signal.name}</span>
              {signal.verified && (
                <CheckCircle className="w-4 h-4 text-blue-500" />
              )}
            </div>
            <span className="text-sm text-muted-foreground">@{signal.handle}</span>
          </div>
          {impactPercent >= 1 && (
            <div className={cn(
              "flex items-center gap-1 font-mono font-bold text-lg",
              isPositiveImpact ? "text-bullish" : "text-bearish"
            )}>
              {isPositiveImpact ? <TrendingUp className="w-5 h-5" /> : <TrendingDown className="w-5 h-5" />}
              {isPositiveImpact ? "+" : "-"}{impactPercent.toFixed(1)}%
            </div>
          )}
        </div>
      </CardHeader>
      <CardContent>
        <p className="text-foreground mb-4 whitespace-pre-wrap">{signal.content}</p>

        {signal.affected_markets && signal.affected_markets.length > 0 && (
          <div className="mb-4 p-3 bg-muted/50 rounded-lg">
            <p className="text-xs text-muted-foreground mb-2">Markets that moved:</p>
            <div className="space-y-1">
              {signal.affected_markets.slice(0, 3).map((market, i) => (
                <MarketMovementBadge key={i} movement={market} />
              ))}
            </div>
          </div>
        )}

        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <div className="flex items-center gap-3">
            <span>{formatTimeAgo(signal.posted_at)}</span>
            {signal.impact_window && (
              <Badge variant="secondary" className="text-xs">
                Impact: {signal.impact_window}
              </Badge>
            )}
          </div>
          <a
            href={signal.tweet_url}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1 text-brand hover:underline"
          >
            View on X <ExternalLink className="w-3.5 h-3.5" />
          </a>
        </div>
      </CardContent>
    </Card>
  );
}

// Market movement badge component
function MarketMovementBadge({ movement }: { movement: MarketMovement }) {
  const isPositive = movement.change >= 0;
  return (
    <div className="flex items-center justify-between text-xs">
      <span className="text-foreground line-clamp-1">{movement.market_title}</span>
      <span className={cn(
        "font-mono font-medium shrink-0 ml-2",
        isPositive ? "text-bullish" : "text-bearish"
      )}>
        {isPositive ? "+" : ""}{(movement.change * 100).toFixed(1)}%
      </span>
    </div>
  );
}

// Section component for displaying signals in articles
interface SocialSignalsSectionProps {
  signals: SocialSignal[];
  title?: string;
}

export function SocialSignalsSection({ signals, title = "Social Signals" }: SocialSignalsSectionProps) {
  if (!signals || signals.length === 0) return null;

  return (
    <section className="mt-8 pt-6 border-t">
      <div className="flex items-center gap-2 mb-4">
        <svg className="w-5 h-5 text-blue-500" fill="currentColor" viewBox="0 0 24 24">
          <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
        </svg>
        <h2 className="font-bold text-lg">{title}</h2>
      </div>
      <div className="space-y-4">
        {signals.map((signal, i) => (
          <SocialSignalCard key={i} signal={signal} />
        ))}
      </div>
    </section>
  );
}

// Inline badges for article headers
interface SignalSourceBadgesProps {
  signals: SocialSignal[];
}

export function SignalSourceBadges({ signals }: SignalSourceBadgesProps) {
  if (!signals || signals.length === 0) return null;

  return (
    <div className="flex flex-wrap gap-2">
      {signals.slice(0, 3).map((signal, i) => (
        <SocialSignalCard key={i} signal={signal} variant="inline" />
      ))}
    </div>
  );
}

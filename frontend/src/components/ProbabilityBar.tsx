import { cn, formatProbability, formatChange } from "@/lib/utils";

interface ProbabilityBarProps {
  probability: number;
  previousProb?: number;
  showChange?: boolean;
  size?: "sm" | "md" | "lg";
  className?: string;
}

export function ProbabilityBar({
  probability,
  previousProb,
  showChange = true,
  size = "md",
  className,
}: ProbabilityBarProps) {
  const change = previousProb !== undefined ? probability - previousProb : undefined;
  const isPositive = change !== undefined && change >= 0;
  const probPercent = probability * 100;

  const sizeClasses = {
    sm: "h-2",
    md: "h-3",
    lg: "h-4",
  };

  return (
    <div className={cn("space-y-1", className)}>
      <div className="flex items-center justify-between text-sm">
        <span className="font-mono font-bold text-lg">
          {formatProbability(probability)}
        </span>
        {showChange && change !== undefined && (
          <span
            className={cn(
              "font-mono font-semibold",
              isPositive ? "text-bullish" : "text-bearish"
            )}
          >
            {formatChange(change)}
          </span>
        )}
      </div>
      <div className={cn("w-full bg-muted rounded-full overflow-hidden", sizeClasses[size])}>
        <div
          className={cn(
            "h-full rounded-full transition-all duration-500",
            probPercent >= 75 ? "bg-bullish" :
            probPercent >= 50 ? "bg-brand" :
            probPercent >= 25 ? "bg-crypto" :
            "bg-bearish"
          )}
          style={{ width: `${probPercent}%` }}
        />
      </div>
      <div className="flex justify-between text-xs text-muted-foreground">
        <span>NO</span>
        <span>YES</span>
      </div>
    </div>
  );
}

// Simple inline probability display
interface ProbabilityInlineProps {
  probability: number;
  change?: number;
  size?: "sm" | "md" | "lg";
}

export function ProbabilityInline({
  probability,
  change,
  size = "md",
}: ProbabilityInlineProps) {
  const isPositive = change !== undefined && change >= 0;

  const sizeClasses = {
    sm: "text-sm",
    md: "text-base",
    lg: "text-xl",
  };

  return (
    <div className={cn("flex items-center gap-2 font-mono", sizeClasses[size])}>
      <span className="font-bold">{formatProbability(probability)}</span>
      {change !== undefined && (
        <span
          className={cn(
            "font-semibold",
            isPositive ? "text-bullish" : "text-bearish"
          )}
        >
          {formatChange(change)}
        </span>
      )}
    </div>
  );
}

// Outcome prices display (for markets with multiple outcomes)
interface OutcomePricesProps {
  outcomes: string[];
  prices: number[];
  size?: "sm" | "md";
}

export function OutcomePrices({ outcomes, prices, size = "md" }: OutcomePricesProps) {
  const sizeClasses = {
    sm: "text-xs gap-2",
    md: "text-sm gap-3",
  };

  return (
    <div className={cn("flex flex-wrap", sizeClasses[size])}>
      {outcomes.map((outcome, index) => (
        <div key={outcome} className="flex items-center gap-1">
          <span className="text-muted-foreground">{outcome}:</span>
          <span className="font-mono font-semibold">
            {prices[index] !== undefined
              ? `${(prices[index] * 100).toFixed(0)}%`
              : "—"}
          </span>
        </div>
      ))}
    </div>
  );
}

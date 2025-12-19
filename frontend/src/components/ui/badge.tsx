import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded-md border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
  {
    variants: {
      variant: {
        default:
          "border-transparent bg-primary text-primary-foreground hover:bg-primary/80",
        secondary:
          "border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80",
        destructive:
          "border-transparent bg-destructive text-destructive-foreground hover:bg-destructive/80",
        outline: "text-foreground",

        // Category variants - matching Polymarket categories
        politics: "border-politics/20 bg-politics/10 text-politics",
        elections: "border-elections/20 bg-elections/10 text-elections",
        crypto: "border-crypto/20 bg-crypto/10 text-crypto",
        finance: "border-finance/20 bg-finance/10 text-finance",
        economy: "border-economy/20 bg-economy/10 text-economy",
        earnings: "border-earnings/20 bg-earnings/10 text-earnings",
        tech: "border-tech/20 bg-tech/10 text-tech",
        sports: "border-sports/20 bg-sports/10 text-sports",
        geopolitics: "border-geopolitics/20 bg-geopolitics/10 text-geopolitics",
        world: "border-world/20 bg-world/10 text-world",
        culture: "border-culture/20 bg-culture/10 text-culture",
        global: "border-global/20 bg-global/10 text-global",

        // Dynamic categories
        trending: "border-trending/20 bg-trending/10 text-trending",
        breaking: "border-breaking/20 bg-breaking/10 text-breaking animate-pulse",
        new: "border-new/20 bg-new/10 text-new",

        // Signal variants
        bullish: "border-bullish/20 bg-bullish/10 text-bullish",
        bearish: "border-bearish/20 bg-bearish/10 text-bearish",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

export { Badge, badgeVariants };

/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  content: [
    './src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}',
  ],
  safelist: [
    // Safelist category colors for dynamic class generation
    {
      pattern: /(bg|text|border)-(politics|elections|crypto|finance|economy|earnings|tech|sports|geopolitics|world|culture|trending|breaking|new|bullish|bearish)/,
      variants: ['hover'],
    },
    {
      pattern: /(bg|text|border)-(politics|elections|crypto|finance|economy|earnings|tech|sports|geopolitics|world|culture|trending|breaking|new|bullish|bearish)\/(10|20|30|50)/,
    },
  ],
  theme: {
    container: {
      center: true,
      padding: "1rem",
      screens: {
        "sm": "640px",
        "md": "768px",
        "lg": "1024px",
        "xl": "1280px",
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        // Brand colors
        brand: {
          DEFAULT: "#0066CC",
          dark: "#004C99",
        },

        // Signal colors
        bullish: "#00A86B",
        bearish: "#DC3545",
        breaking: "#FF4757",
        trending: "#FF6B6B",
        new: "#22C55E",

        // Category colors - matching Polymarket
        politics: "#6B46C1",
        elections: "#7C3AED",
        crypto: "#F7931A",
        finance: "#3B82F6",
        economy: "#0EA5E9",
        earnings: "#6366F1",
        tech: "#0891B2",
        sports: "#10B981",
        geopolitics: "#8B5CF6",
        world: "#EC4899",
        culture: "#F43F5E",
        global: "#059669",

        // UI colors
        border: "#E5E5E5",
        input: "#E5E5E5",
        ring: "#0066CC",
        background: "#FFFFFF",
        foreground: "#1A1A1A",
        primary: {
          DEFAULT: "#0066CC",
          foreground: "#FFFFFF",
        },
        secondary: {
          DEFAULT: "#F5F5F5",
          foreground: "#1A1A1A",
        },
        muted: {
          DEFAULT: "#F5F5F5",
          foreground: "#666666",
        },
        accent: {
          DEFAULT: "#F5F5F5",
          foreground: "#1A1A1A",
        },
        destructive: {
          DEFAULT: "#DC3545",
          foreground: "#FFFFFF",
        },
        card: {
          DEFAULT: "#FFFFFF",
          foreground: "#1A1A1A",
        },
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "sans-serif"],
        mono: ["JetBrains Mono", "monospace"],
        serif: ["Georgia", "serif"],
      },
      borderRadius: {
        lg: "0.5rem",
        md: "0.375rem",
        sm: "0.25rem",
      },
      keyframes: {
        "pulse-slow": {
          "0%, 100%": { opacity: 1 },
          "50%": { opacity: 0.5 },
        },
        "pulse-live": {
          "0%, 100%": { opacity: 1 },
          "50%": { opacity: 0.4 },
        },
        "slide-in": {
          "0%": { transform: "translateY(-10px)", opacity: 0 },
          "100%": { transform: "translateY(0)", opacity: 1 },
        },
      },
      animation: {
        "pulse-slow": "pulse-slow 2s cubic-bezier(0.4, 0, 0.6, 1) infinite",
        "pulse-live": "pulse-live 1.5s ease-in-out infinite",
        "slide-in": "slide-in 0.3s ease-out",
      },
    },
  },
  plugins: [],
}

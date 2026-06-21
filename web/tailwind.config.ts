import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        bg: "var(--bg)",
        surface: "var(--surface)",
        "surface-2": "var(--surface-2)",
        ink: "var(--ink)",
        "ink-muted": "var(--ink-muted)",
        accent: "var(--accent)",
        "accent-2": "var(--accent-2)",
        border: "var(--border)",
        danger: "var(--danger)",
        success: "var(--success)",
        warning: "var(--warning)",
      },
      fontFamily: {
        sans: ["var(--font-inter)", "system-ui", "sans-serif"],
        mono: ["ui-monospace", "SFMono-Regular", "Menlo", "monospace"],
      },
      borderRadius: {
        DEFAULT: "4px",
      },
      transitionTimingFunction: {
        "ease-out-soft": "cubic-bezier(0.23, 1, 0.32, 1)",
      },
      boxShadow: {
        glow: "0 0 40px -8px var(--accent-glow)",
        card: "0 8px 30px -12px rgba(0,0,0,0.5)",
        lift: "0 18px 50px -20px rgba(83,58,253,0.45)",
      },
      keyframes: {
        "fade-up": {
          "0%": { opacity: "0", transform: "translateY(8px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
      },
      animation: {
        "fade-up": "fade-up 250ms cubic-bezier(0.25,0.46,0.45,0.94) both",
      },
    },
  },
  plugins: [],
};

export default config;

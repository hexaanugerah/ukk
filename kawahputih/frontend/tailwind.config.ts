import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: "class",
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}", "./features/**/*.{ts,tsx}"],
  theme: {
    screens: {
      sm: "640px",
      md: "768px",
      lg: "1024px",
      xl: "1280px",
      "2xl": "1536px",
    },
    extend: {
      colors: {
        // Palet Part 2.4 — dinamai brand (bukan primary) agar tidak memicu rename besar di komponen lama. Nilai token persis seperti spesifikasi.
        brand: {
          light: "#14B8A6",
          DEFAULT: "#0F766E",
          dark: "#115E59",
          50: "#f0fdfa",
          100: "#ccfbf1",
          500: "#14B8A6",
          600: "#0F766E",
          700: "#115E59",
        },
        secondary: { DEFAULT: "#0284C7" },
        accent: { DEFAULT: "#F59E0B" },
        success: "#16A34A",
        warning: "#EAB308",
        danger: "#DC2626",
        info: "#2563EB",
        surface: { DEFAULT: "#FFFFFF", dark: "#0F172A" },
        appbg: { DEFAULT: "#F8FAFC", dark: "#020617" },
        borderc: { DEFAULT: "#E2E8F0", dark: "#1E293B" },
        text: {
          primary: "#0F172A",
          secondary: "#475569",
          muted: "#94A3B8",
          "primary-dark": "#F8FAFC",
        },
      },
      fontFamily: {
        heading: ["var(--font-poppins)", "sans-serif"],
        body: ["var(--font-inter)", "sans-serif"],
      },
      fontSize: {
        display: ["64px", { lineHeight: "1.1", fontWeight: "700" }],
        hero: ["56px", { lineHeight: "1.1", fontWeight: "700" }],
        h1: ["48px", { lineHeight: "1.15", fontWeight: "700" }],
        h2: ["40px", { lineHeight: "1.2", fontWeight: "700" }],
        h3: ["32px", { lineHeight: "1.25", fontWeight: "700" }],
        h4: ["28px", { lineHeight: "1.3", fontWeight: "700" }],
        h5: ["24px", { lineHeight: "1.35", fontWeight: "700" }],
        h6: ["20px", { lineHeight: "1.4", fontWeight: "700" }],
        "body-xl": ["18px", { lineHeight: "1.6" }],
        "body-sm": ["14px", { lineHeight: "1.5" }],
        caption: ["12px", { lineHeight: "1.4" }],
      },
      borderRadius: {
        card: "6px",
        btn: "6px",
        input: "6px",
        dialog: "8px",
      },
      boxShadow: {
        floating: "0 25px 50px -12px rgb(0 0 0 / 0.25)",
      },
      transitionDuration: {
        150: "150ms",
        250: "250ms",
        300: "300ms",
        500: "500ms",
      },
    },
  },
  plugins: [],
};

export default config;

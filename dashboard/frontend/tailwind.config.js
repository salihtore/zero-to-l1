/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        bg: {
          DEFAULT: "#0A0A0A",
          surface: "#141414",
          surfaceHover: "#1A1A1A",
          elevated: "#1F1F1F",
        },
        border: {
          DEFAULT: "#262626",
          light: "#2E2E2E",
        },
        accent: {
          DEFAULT: "#E84142",
          hover: "#F2565A",
          muted: "#E8414220",
        },
        text: {
          primary: "#F5F5F5",
          secondary: "#A3A3A3",
          tertiary: "#6B6B6B",
        },
        status: {
          online: "#3FB950",
          onlineMuted: "#3FB95020",
          pending: "#D29922",
          pendingMuted: "#D2992220",
          offline: "#6B6B6B",
          offlineMuted: "#6B6B6B20",
        },
      },
      fontFamily: {
        sans: [
          "Inter",
          "-apple-system",
          "BlinkMacSystemFont",
          "Segoe UI",
          "Roboto",
          "sans-serif",
        ],
        mono: ["JetBrains Mono", "SFMono-Regular", "Menlo", "monospace"],
      },
    },
  },
  plugins: [],
};
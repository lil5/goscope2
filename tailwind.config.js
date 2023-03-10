/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["admin.html"],
  theme: {
    extend: {},
  },
  daisyui: {
    themes: [
      {
        light: {
          ...require("daisyui/src/colors/themes")["[data-theme=light]"],
          primary: "#0EA5E9",
          secondary: "#14B8A6",
          accent: "#EAB308",
          info: "#06B6D4",
          "info-content": "#FAFAFA",
          success: "#16A34A",
          "success-content": "#FAFAFA",
          warning: "#F59E0B",
          "warning-content": "#FAFAFA",
          error: "#DC2626",
          "error-content": "#FAFAFA",
        },
      },
      {
        dark: {
          ...require("daisyui/src/colors/themes")["[data-theme=dark]"],
          primary: "#0EA5E9",
          secondary: "#14B8A6",
          accent: "#EAB308",
          "neutral-content": "#FAFAFA",
          "base-content": "#FAFAFA",
          // "base-100": "#FAFAFA",
          info: "#06B6D4",
          "info-content": "#FAFAFA",
          success: "#16A34A",
          "success-content": "#FAFAFA",
          warning: "#F59E0B",
          "warning-content": "#FAFAFA",
          error: "#DC2626",
          "error-content": "#FAFAFA",
        },
      },
    ],
  },
  plugins: [require("daisyui")],
};

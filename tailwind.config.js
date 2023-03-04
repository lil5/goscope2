/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["admin.html"],
  theme: {
    extend: {},
  },
  daisyui: {
    themes: ["light", "dark"],
  },
  plugins: [require("daisyui")],
};

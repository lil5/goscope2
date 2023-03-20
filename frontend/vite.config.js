import { createHtmlPlugin } from "vite-plugin-html";

/** @type {import('vite').UserConfig} */
export default {
  base: "./",
  appType: "mpa",
  build: {
    assetsDir: ".",
    rollupOptions: {
      output: {
        entryFileNames: `[name].js`,
        chunkFileNames: `[name].js`,
        assetFileNames: `[name].[ext]`,
      },
    },
  },
  server: {
    proxy: {
      "/api": {
        target: "http://127.0.0.1:8080/goscope2/api",
        auth: "admin:admin",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
    },
  },
  plugins: [
    createHtmlPlugin({
      minify: true,
      /**
       * After writing entry here, you will not need to add script tags in `index.html`, the original tags need to be deleted
       * @default src/main.ts
       */
      entry: "main.js",
      /**
       * If you want to store `index.html` in the specified folder, you can modify it, otherwise no configuration is required
       * @default index.html
       */
      template: "index.html",
      inject: {
        tags: [
          {
            injectTo: "body",
            tag: "script",
            attrs: {
              defer: "defer",
              src: "https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js",
            },
          },
          {
            injectTo: "head",
            tag: "link",
            attrs: {
              href: "https://unpkg.com/css.gg/icons/all.css",
              rel: "stylesheet",
            },
          },
          {
            injectTo: "head",
            tag: "script",
            attrs: { src: "https://cdn.jsdelivr.net/npm/dayjs@1/dayjs.min.js" },
          },
          {
            injectTo: "head",
            tag: "script",
            attrs: {
              src: "https://cdn.jsdelivr.net/npm/dayjs@1/plugin/calendar.js",
            },
          },
          {
            injectTo: "head",
            tag: "script",
            attrs: {
              src: "https://cdn.jsdelivr.net/npm/dayjs@1/plugin/relativeTime.js",
            },
          },
          {
            injectTo: "head",
            tag: "script",
            attrs: {
              src: "https://cdn.jsdelivr.net/npm/dayjs@1/plugin/updateLocale.js",
            },
          },
        ],
      },
    }),
  ],
};

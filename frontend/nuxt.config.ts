import { fileURLToPath } from "node:url";
import vuetify from "vite-plugin-vuetify";
import { dockerPagesWatchPlugin } from "./plugins-dev/docker-pages-watch";

const rootDir = fileURLToPath(new URL(".", import.meta.url));

export default defineNuxtConfig({
  ssr: false,

  future: {
    compatibilityVersion: 3,
  },

  typescript: {
    shim: false,
  },

  // SPA only: devtools add a lot of dev-server startup overhead.
  devtools: { enabled: false },

  runtimeConfig: {
    apiUrl: process.env.API_URL || "http://localhost:8080",
    public: {
      apiUrl: process.env.NUXT_PUBLIC_API_URL || "",
      siteUrl: process.env.NUXT_PUBLIC_SITE_URL || "http://localhost",
      turnstileSiteKey: process.env.NUXT_PUBLIC_TURNSTILE_SITE_KEY || "",
      // Cookie Secure flag for auth tokens. Must be false on plain HTTP (local).
      cookieSecure:
        process.env.NUXT_PUBLIC_COOKIE_SECURE === "true" ||
        process.env.NUXT_PUBLIC_COOKIE_SECURE === "1",
      // Dev-only login autofill (never exposed when APP_ENV is not development).
      devLoginEmail:
        process.env.APP_ENV === "development"
          ? process.env.SUPERADMIN_EMAIL || ""
          : "",
      devLoginPassword:
        process.env.APP_ENV === "development"
          ? process.env.SUPERADMIN_PASSWORD || ""
          : "",
    },
  },

  hooks: {
    // Required in dev: Vite's dep-scan reads .nuxt/*.mjs from disk before
    // Nuxt would lazily write virtual templates.
    "app:templates"(app) {
      for (const template of app.templates) {
        template.write = true;
      }
      app.templates.push({
        filename: "plugins.mjs",
        write: true,
        getContents: () => "export { default } from './plugins.client.mjs'\n",
      });
    },
    "vite:extendConfig"(config, { isClient }) {
      if (!isClient) {
        return;
      }

      // Vite runs in middlewareMode under Nitro, so HMR must use its own
      // websocket port. nginx proxies /_nuxt/hmr/ -> web-dev:24678.
      config.server ??= {};
      config.server.allowedHosts = true;
      // Vite 6 denies transform for URLs matching /\.svg\b/ unless fs.strict
      // is off or the path is allow-listed. Nuxt's virtual:public?%2F…svg
      // IDs contain ".svg" but are not real files, so they 404 and blank
      // the SPA whenever a template references a public SVG.
      config.server.fs = {
        ...(config.server.fs ?? {}),
        strict: false,
      };
      config.server.hmr = {
        protocol: "ws",
        port: 24678,
        clientPort: 80,
        path: "hmr/",
      };

      // Tree-shaken Vuetify (components/directives imported on demand).
      config.plugins ??= [];
      config.plugins.push(vuetify({ autoImport: true }));
    },
  },

  // Nuxt's own watcher: same polling, minus heavy directories.
  watchers: {
    chokidar: {
      usePolling: true,
      interval: 1000,
      ignored: ["**/node_modules/**", "**/.nuxt/**", "**/.output/**", "**/.data/**"],
    },
  },

  devServer: {
    host: "0.0.0.0",
    port: 3000,
  },

  vite: {
    server: {
      host: "0.0.0.0",
      allowedHosts: true,
      fs: {
        strict: false,
      },
      watch: {
        // Bind mounts on Windows/macOS emit no inotify events, so poll, but
        // only the source files that matter and not too aggressively.
        usePolling: true,
        interval: 1000,
        binaryInterval: 3000,
        ignored: [
          "**/node_modules/**",
          "**/.nuxt/**",
          "**/.output/**",
          "**/.data/**",
          "**/.git/**",
        ],
      },
    },
    optimizeDeps: {
      exclude: ["nuxt"],
      // Pre-bundle heavy deps at startup so Vite does not discover them
      // lazily and trigger optimize + full-reload cycles on first page load.
      include: [
        "vuetify",
        "apexcharts",
        "vue3-apexcharts",
        "vue-tabler-icons",
        "vue3-perfect-scrollbar",
        "@iconify/vue",
      ],
    },
    plugins: [dockerPagesWatchPlugin(`${rootDir}/pages`)],
  },

  app: {
    head: {
      title: "Company Profile",
    },
  },

  build: {
    transpile: ["vuetify"],
  },

  nitro: {
    serveStatic: true,
  },

  sourcemap: { server: false, client: false },
  compatibilityDate: "2024-04-03",
});

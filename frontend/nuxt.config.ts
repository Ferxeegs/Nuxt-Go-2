import { fileURLToPath } from "node:url";
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

  // Inline module: hard-restart when pages are added/removed so Vite does not
  // keep serving a stale `.nuxt/routes.mjs` transform (common on Docker Desktop).
  modules: [
    (_options, nuxt) => {
      if (!nuxt.options.dev) return;

      let restarting = false;
      nuxt.hook("builder:watch", async (event, rawPath) => {
        if (restarting) return;
        if (event !== "add" && event !== "unlink") return;

        const path = String(rawPath).replace(/\\/g, "/");
        if (!path.includes("/pages/") || !path.endsWith(".vue")) return;

        restarting = true;
        console.info(
          `[docker-pages-restart] page ${event}: ${path} — hard restart to refresh routes`,
        );
        try {
          await nuxt.callHook("restart", { hard: true });
        } finally {
          setTimeout(() => {
            restarting = false;
          }, 5000);
        }
      });
    },
  ],

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

      config.plugins ??= [];
      config.plugins.push(dockerPagesWatchPlugin(`${rootDir}/pages`));
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
        usePolling: true,
        // Faster detection of newly created page files on Docker Desktop.
        interval: 500,
        binaryInterval: 1000,
      },
    },
    optimizeDeps: {
      exclude: ["nuxt"],
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

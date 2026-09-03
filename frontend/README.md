# Frontend — Nuxt 3 Company Profile

Nuxt 3 + Vuetify admin shell (based on the Spike free template). This is the starting UI for the company-profile boilerplate; public landing pages, auth, and API wiring are not implemented yet.

The app currently runs as an **SPA** (`ssr: false` in `nuxt.config.ts`). That is fine for the admin shell. Turn SSR on when you add public, SEO-sensitive pages.

## Scripts

```bash
npm install
npm run dev       # nuxt dev — http://localhost:3000
npm run build     # production Nitro server in .output/
npm run preview   # preview the production build
```

## Environment

All env vars live in the **repo-root** `.env` (see `../.env.example`). Nuxt reads them via `runtimeConfig`:

| Variable | Scope | Purpose |
|---|---|---|
| `API_URL` | server-only | Go API origin for Nitro / server routes. Compose sets `http://api-dev:8080` or `http://api:8080`. |
| `NUXT_PUBLIC_API_URL` | public | Browser API origin. Prefer same-origin (`http://localhost`) because nginx proxies `/api/` to Go. |
| `NUXT_PUBLIC_SITE_URL` | public | Canonical site URL. |
| `NUXT_PUBLIC_TURNSTILE_SITE_KEY` | public | Cloudflare Turnstile site key. |

Access in code with `useRuntimeConfig()`. Rebuild the production `web` image after changing `NUXT_PUBLIC_*`.

## Docker

The usual workflow is from the repo root:

```bash
make dev    # Nuxt in web-dev, nginx on :80
make prod   # Nitro server in web, nginx-prod on :8080
```

- Dev image: `Dockerfile.dev` (`nuxt dev --host 0.0.0.0`).
- Prod image: `Dockerfile` (multi-stage, runs `node .output/server/index.mjs` as `node`).
- Vite polling is enabled for Docker Desktop bind mounts on Windows/macOS.

## Host-only run

```bash
cd frontend
npm install
npm run dev
```

Point `API_URL` at a reachable Go process and keep `ALLOWED_ORIGINS` in sync with the browser origin.

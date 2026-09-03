# Go-Nuxt Company Profile Boilerplate (Monorepo)

Starter kit for company profile / landing page projects: a Go (Gin) headless REST API and a Nuxt 3 frontend in one repo, both deployed as Docker containers on the same server. Nginx is the only public entry; the Go API is internal to the Docker network except `/api/` and `/uploads/`.

```
nuxt-go/
├── backend/                 # Go + Gin + GORM + MySQL API
├── frontend/                # Nuxt 3 app (public site + admin)
├── deploy/nginx/            # nginx snippets (dev / prod container / global TLS)
├── docker-compose.yml       # profiles: dev (api+web+nginx) / prod (api+web+nginx-prod)
├── .env.example
└── Makefile
```

## Quick start

Prerequisites: Docker, and a **shared MySQL + shared Redis** container already running and reachable (see "Shared infrastructure" below). This stack does not run its own MySQL/Redis containers.

1. Copy the env template and fill in real values:
   ```bash
   cp .env.example .env
   ```
2. Make sure the shared network exists and your shared MySQL/Redis containers are attached to it:
   ```bash
   make network   # creates "app-bridge" if missing
   ```
3. Run in development (Go via [air](https://github.com/air-verse/air), Nuxt via `nuxt dev`, nginx on port 80):
   ```bash
   make dev
   # equivalent to: docker compose --profile dev up --build
   ```
   Open http://localhost — nginx proxies `/` to Nuxt, `/api/` to Go, and `/uploads/` to Go.
4. Run in production (`api` + `web` + `nginx-prod` on port 8080):
   ```bash
   make prod
   # equivalent to: docker compose --profile prod up --build -d
   ```
   Open http://localhost:8080 (or your server IP). `api` and `web` are not published to the host.

   Point the server's global nginx (TLS) at `http://127.0.0.1:8080`. See [`deploy/nginx/ponpes-kebangsaan.global.conf`](deploy/nginx/ponpes-kebangsaan.global.conf) (`template-nginx`).

   Set `NUXT_PUBLIC_SITE_URL` / `NUXT_PUBLIC_API_URL` to your public origin (e.g. `https://ponpes.ferxcode.my.id`) in `.env` and rebuild `web` after changing them.

On first start the app auto-migrates the schema. Seeding is controlled by `DB_SEED` (default **off** in production, **on** in development):

- `DB_SEED=true` — roles, permissions, SuperAdmin, default web settings

First production deploy on an empty database: set `DB_SEED=true`, start once, then set `DB_SEED=false` and recreate the `api` container.

When seeding runs, it creates:
- Roles: `SuperAdmin`, `Admin`, `Editor`, `Operator`
- Permissions for core modules (`user.*`, `role.*`, `setting.*`, `media.*`, `activity.*`, `invoice.*`)
- A `SuperAdmin` account from `SUPERADMIN_EMAIL` / `SUPERADMIN_PASSWORD` — **change the password after first login**.
- Default `web_settings` rows.

API base path: `/api/v1`. Health checks: `GET /healthz` (liveness), `GET /readyz` (DB reachability). Swagger UI (non-production only): `GET /swagger/index.html`.

## Shared infrastructure

MySQL and Redis are **not** part of this stack — they're shared containers used across multiple company-profile projects:

- `DB_HOST` / `REDIS_HOST` in `.env` must point at the shared containers' service names (e.g. `shared-mysql`, `shared-redis`).
- `DB_NAME` must be **unique per project** (shared MySQL instance, separate schemas).
- `REDIS_PREFIX` must be **unique per project** (shared Redis instance; namespaces rate-limit/blacklist keys so projects don't collide).
- `docker-compose.yml` joins an `external` Docker network (`app-bridge` by default) — create it once with `docker network create app-bridge` and make sure the shared MySQL/Redis containers are attached to it.

If you ever run this project standalone without shared infra, add local `mysql`/`redis` services back into `docker-compose.yml` and point `DB_HOST`/`REDIS_HOST` at them instead.

## Environment variables

See `.env.example` for the full list with defaults (backend + frontend in one file). Key groups:
- **App**: `APP_ENV`, `APP_PORT`
- **Database**: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`, `DB_SEED`
- **Redis**: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`, `REDIS_PREFIX`
- **Auth**: `JWT_SECRET`, `JWT_ACCESS_TTL`, `JWT_REFRESH_TTL`, `BCRYPT_COST`, `SUPERADMIN_EMAIL`, `SUPERADMIN_PASSWORD`, `TURNSTILE_SECRET_KEY` (required in production; verifies Cloudflare Turnstile on login)
- **CORS**: `ALLOWED_ORIGINS` (comma-separated whitelist, no wildcard)
- **Rate limiting**: `RATE_LIMIT_PUBLIC_PER_MINUTE`, `RATE_LIMIT_LOGIN_PER_MINUTE`
- **Upload**: `MAX_UPLOAD_SIZE_MB`, `UPLOAD_DIR`
- **Email (optional)**: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM`, `NOTIFY_EMAIL`
- **Frontend**: `API_URL` (server-only; Compose overrides it to the internal Go service), `NUXT_PUBLIC_API_URL`, `NUXT_PUBLIC_SITE_URL`, and `NUXT_PUBLIC_TURNSTILE_SITE_KEY` (injected via Nuxt `runtimeConfig`)

Startup fails fast (with a clear error) if any required variable (`DB_HOST`, `DB_USER`, `DB_NAME`, `JWT_SECRET`, `SUPERADMIN_EMAIL`, `SUPERADMIN_PASSWORD`) is missing. In production, `TURNSTILE_SECRET_KEY` is also required.

## Architecture

Clean/modular layering per feature, under `backend/internal/modules/<name>/`:

```
Handler (Gin controller: bind + validate input, call service, format response)
   -> Service (business logic, RBAC-specific rules, orchestration)
      -> Repository (GORM data access, behind an interface)
         -> Domain (GORM struct/entity, shared across modules)
```

Dependencies always point inward (Handler -> Service -> Repository -> Domain), so services/handlers can be unit tested by mocking the Repository interface (see `backend/internal/modules/auth/service_test.go`).

### Adding a new module

1. Add/extend a domain struct in `internal/domain/`.
2. Add it to `internal/database/migration.go`'s `AutoMigrate` call (respecting FK order).
3. Create `internal/modules/<name>/{dto,repository,service,handler}.go` following an existing module (e.g. `websetting` is a simple settings CRUD example).
4. Wire the handler into `router.Handlers` in `internal/router/router.go`, build it in `cmd/api/main.go`'s `buildHandlers`, and mount routes in `internal/router/public_routes.go` / `admin_routes.go` with the right `middleware.RequirePermission(...)`.
5. Add any new permission keys to `internal/constants/permissions.go` (and to `RolePermissionMap` if a default role should have them) so the seeder grants them.

## Security

- JWT access tokens (short-lived) + rotating refresh tokens (hashed at rest, revoked on logout/rotation) — see `internal/modules/auth`.
- RBAC via `middleware.RequirePermission("resource.action")` per endpoint.
- CORS whitelist only (`ALLOWED_ORIGINS`), no wildcard.
- Redis-backed rate limiting (works across multiple app instances), strict on `/auth/login` and `/public/contact`.
- Security headers (`X-Content-Type-Options`, `X-Frame-Options`, HSTS, CSP) via `internal/middleware/security_headers.go`.
- Upload validation: extension whitelist + sniffed MIME type (not just the client-supplied one), size cap.
- Article/rich-text HTML sanitized with `bluemonday` before storage (stored-XSS prevention).
- Every upload recorded in `media_assets` so orphaned files can be audited/cleaned up later.

## Testing & linting

```bash
make test   # go test ./... -cover
make lint   # go vet + gofmt -l
```

CI (`.github/workflows/ci.yml`) runs backend lint/test/image build, a Nuxt production build, and `docker compose config` validation on every push/PR that touches the relevant paths.

## API documentation

Swagger/OpenAPI is generated from handler annotations via `swaggo/swag`. To regenerate after changing/adding endpoints:

```bash
cd backend
go run github.com/swaggo/swag/cmd/swag@v1.8.12 init -g cmd/api/main.go -o docs --parseDependency --parseInternal
```

(Pin the CLI to `v1.8.12` to match the `github.com/swaggo/swag` library version in `go.mod` — newer CLI versions emit fields the older library struct doesn't have.)

## Known trade-offs

- Final production image is ~60MB (binary ~32MB), above the PRD's 30MB aspirational target — the feature set (GORM+MySQL driver, validator, bluemonday HTML sanitizer, embedded Swagger spec) accounts for the size. Trim further with UPX or by dropping embedded Swagger docs from the prod build if a stricter budget is required.
- `go.mod` targets Go 1.25 (PRD asked for "1.22+"): the current release lines of `gin-gonic/gin`, `spf13/viper`, `swaggo/gin-swagger` and `redis/go-redis/v9` all require >=1.23-1.25 toolchains.

## Frontend

`frontend/` — Nuxt 3 + Vuetify (Spike admin template as the starting UI), consuming this backend as a headless API. See [`frontend/README.md`](frontend/README.md). Env vars live in the root `.env` (`API_URL`, `NUXT_PUBLIC_API_URL`, `NUXT_PUBLIC_SITE_URL`).

The usual workflow is `make dev` / `make prod` (frontend runs in a container). To run Nuxt on the host instead:

```bash
cd frontend
npm install
npm run dev
```

Point `API_URL` at `http://localhost:8080` (only works if the Go port is published) and keep `ALLOWED_ORIGINS` in sync with the browser origin. Browser traffic to `/api/` and `/uploads` is same-origin when nginx is in front.

The template currently runs as an SPA (`ssr: false`). Enable SSR later when public company-profile pages need SEO.

# PRD — Boilerplate Backend Go (Gin) untuk Website Company Profile / Landing Page Dinamis

| | |
|---|---|
| **Nama Proyek** | Go-Gin Company Profile Boilerplate (`go-cp-boilerplate`) |
| **Versi Dokumen** | 1.1 |
| **Tanggal** | 3 September 2026 |
| **Tipe Dokumen** | Product Requirements Document (PRD) |
| **Target Pengguna** | Tim internal / developer yang akan reuse boilerplate untuk proyek company profile / landing page klien |

---

## 1. Latar Belakang & Tujuan

Banyak proyek website company profile atau landing page berskala kecil–menengah (bukan sistem enterprise besar seperti SIMAK ALMA) yang tetap membutuhkan backend dinamis untuk mengelola konten (artikel, layanan, pengaturan web, pesan kontak) tanpa harus membangun dari nol setiap kali ada proyek baru.

**Tujuan boilerplate ini:**

1. Menyediakan starter kit backend Go + Gin yang **reusable** untuk proyek company profile/landing page.
2. Frontend (**Nuxt 3**, di-deploy sebagai container Docker di belakang nginx yang sama) mengonsumsi backend ini sebagai headless API.
3. Backend sudah mengikuti **standar keamanan dan arsitektur industri** (clean layering, RBAC, JWT, rate limiting, dsb) sehingga tidak perlu "mengulang dari nol" aspek keamanan dasar di tiap proyek.
4. Mudah di-deploy via Docker ke VPS/Cloud dengan effort minimal (`docker compose up`).

**Out of scope (bukan tujuan boilerplate ini):**
- E-commerce (payment gateway, cart, dsb) — jika dibutuhkan, jadi modul tambahan terpisah.
- Multi-tenant SaaS.
- Real-time feature (websocket/chat).

---

## 2. Tech Stack

| Kategori | Pilihan | Alasan |
|---|---|---|
| Bahasa | Go 1.22+ | Performa tinggi, binary kecil, cocok untuk container ringan |
| Web Framework | Gin | Ringan, cepat, ekosistem middleware luas |
| ORM | GORM | Auto-migration, mendukung MySQL, relasi mudah didefinisikan |
| Database | MySQL 8 | Sesuai kebutuhan di poin 7 (volume MySQL di docker-compose) |
| Config | Viper + godotenv | Viper untuk config terstruktur (yaml/env), godotenv untuk load `.env` lokal |
| Auth Token | JWT (`golang-jwt/jwt/v5`) | Standar de facto untuk stateless auth |
| Password Hashing | bcrypt (`golang.org/x/crypto/bcrypt`) | Standar industri |
| Validasi | `go-playground/validator` (built-in Gin binding) | Terintegrasi langsung dengan Gin binding |
| Logging | `zerolog` atau `zap` | Structured logging, performa tinggi |
| API Docs | `swaggo/swag` (Swagger/OpenAPI) | Auto-generate docs dari comment annotation |
| Cache/Rate limit store | Redis (**shared container**, eksternal) | Rate limiting terdistribusi & token blacklist, memakai instance Redis bersama — bukan container Redis milik proyek ini |
| Container | Docker + Docker Compose | Sesuai requirement |
| CI/CD | GitHub Actions | Standar industri untuk lint/test/build otomatis |

> **Catatan:** Baik MySQL maupun Redis **tidak** dijalankan sebagai container di dalam boilerplate ini — keduanya berupa **shared infrastructure** yang sudah berjalan terpisah dan diakses lewat Docker external network (lihat Bab 10.2). Jika suatu saat proyek berjalan sendiri tanpa shared infra, cukup tambahkan kembali service `mysql`/`redis` di `docker-compose.yml` dan arahkan `DB_HOST`/`REDIS_HOST` ke service lokal tersebut.

---

## 3. Struktur Folder Proyek (Clean/Modular Architecture)

```
go-cp-boilerplate/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point aplikasi
├── internal/
│   ├── config/                     # Load & parse .env / config.yaml (Viper)
│   ├── database/
│   │   ├── connection.go           # Koneksi GORM ke MySQL
│   │   ├── migration.go            # Auto migration
│   │   └── seeder.go               # Seeder SuperAdmin, Role, Web Settings
│   ├── domain/                     # Model/Entity (struct GORM)
│   │   ├── user.go
│   │   ├── role.go
│   │   ├── permission.go
│   │   ├── role_permission.go
│   │   ├── web_setting.go
│   │   ├── article.go
│   │   ├── service.go
│   │   ├── contact_message.go
│   │   └── media_asset.go
│   ├── modules/                    # Fitur, per-modul (Handler → Service → Repository)
│   │   ├── auth/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── dto.go
│   │   ├── user/
│   │   ├── role/
│   │   ├── article/
│   │   ├── servicecontent/         # nama "service" bentrok kata kunci Go, jadi servicecontent
│   │   ├── contactmessage/
│   │   ├── websetting/
│   │   └── media/
│   ├── middleware/
│   │   ├── auth_middleware.go      # JWT validator
│   │   ├── rbac_middleware.go      # Role/permission checker
│   │   ├── cors_middleware.go
│   │   ├── rate_limiter.go
│   │   ├── security_headers.go
│   │   ├── logger_middleware.go    # Structured request logging
│   │   └── recovery_middleware.go  # Global panic handler
│   ├── router/
│   │   ├── router.go                # Entry setup semua route
│   │   ├── public_routes.go
│   │   └── admin_routes.go
│   ├── pkg/
│   │   ├── response/                # Standardized JSON response helper
│   │   ├── jwtutil/                 # Generate & validasi token
│   │   ├── hash/                    # bcrypt wrapper
│   │   ├── pagination/              # Helper pagination
│   │   ├── validator/               # Custom validator rules
│   │   ├── uploader/                # Upload handler, MIME check
│   │   └── logger/                  # Init zerolog/zap
│   └── constants/                   # Enum, error codes, permission keys
├── docs/                            # Swagger auto-generated docs
├── storage/
│   └── uploads/                     # Direktori upload (volume Docker)
├── migrations/                      # (opsional) SQL migration manual, jika tak pakai auto-migrate saja
├── scripts/
│   └── wait-for-db.sh
├── .env.example
├── .air.toml                        # Hot reload dev (opsional, best practice DX)
├── Dockerfile                       # Multi-stage (target dev / prod)
├── go.mod / go.sum
└── README.md                        # Dokumentasi backend ada di README root monorepo
```

`docker-compose.yml` dan `Makefile` berada di **root monorepo**, bukan di dalam `backend/`.

**Prinsip layering per modul:**

```
Handler (Gin Controller)
   ↓  menerima request, validasi input, panggil service
Service (Usecase / Business Logic)
   ↓  logika bisnis, orkestrasi, aturan RBAC spesifik fitur
Repository (Data Access Layer)
   ↓  query ke database via GORM
Domain (Struct/Entity)
```

Dependency selalu mengarah ke dalam (Handler → Service → Repository → Domain), memudahkan unit testing dengan mocking di layer Repository/Service (pakai interface).

---

## 4. Struktur Database & Model (GORM)

### 4.1 Modul Autentikasi & RBAC

**`users`**
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | uint (PK) | |
| name | varchar | |
| email | varchar, unique | |
| password | varchar | hash bcrypt |
| role_id | uint (FK → roles.id) | |
| avatar | varchar, nullable | path media_assets |
| is_active | boolean | default true |
| last_login_at | datetime, nullable | |
| created_at / updated_at / deleted_at | timestamp | soft delete GORM |

**`roles`**
| id | name (SuperAdmin, Admin, Editor) | description | created_at/updated_at |

**`permissions`**
| id | name (contoh: `article.create`, `user.delete`) | description |

**`role_permissions`** (pivot many-to-many)
| id | role_id (FK) | permission_id (FK) |

### 4.2 Modul Konten Umum

**`web_settings`** — didesain sebagai key-value agar fleksibel tanpa migrasi tiap kali ada field baru:
| id | key (unique, contoh: `site_name`, `logo`, `contact_email`, `contact_phone`, `social_facebook`, `footer_text`) | value (text) | type (string/text/json/image) | updated_at |

**`articles`**
| id | title | slug (unique) | excerpt | content (longtext) | thumbnail | status (draft/published) | author_id (FK users) | published_at | created_at/updated_at/deleted_at |

**`services`** (portofolio/layanan)
| id | title | slug | description | icon/image | order (untuk sorting tampilan) | is_active | created_at/updated_at |

**`contact_messages`**
| id | name | email | phone | subject | message | is_read (boolean) | created_at |

### 4.3 Modul Media/File

**`media_assets`**
| id | file_name | path | mime_type | size (bytes) | uploaded_by (FK users, nullable) | related_module (contoh: "article", "user_avatar") | related_id (nullable) | created_at |

> Tujuan tabel ini: setiap upload tercatat sehingga file lama/orphan mudah dibersihkan via cron/job, dan menghindari file "menumpuk" tanpa referensi.

### 4.4 ERD (ringkas, hubungan antar tabel)

```
roles ──1:N── users ──1:N── articles
  │
  1:N
  │
role_permissions ──N:1── permissions

users ──1:N── media_assets (uploaded_by)
articles/services ──referensi opsional── media_assets

web_settings   : standalone (key-value)
contact_messages : standalone
```

---

## 5. Sistem Autentikasi & Otorisasi

| Fitur | Spesifikasi |
|---|---|
| Password Hashing | bcrypt cost factor 10–12, konfigurasi via env |
| Access Token (JWT) | Payload: `user_id`, `role`, `permissions[]` (bisa di-embed atau di-lookup ulang), `exp`. Masa aktif: 15–60 menit (configurable) |
| Refresh Token | Disimpan hashed di DB (tabel `refresh_tokens`) atau Redis, masa aktif 7 hari, dengan rotasi token (invalidate token lama setiap refresh) untuk mencegah replay attack |
| Middleware Auth | Membaca header `Authorization: Bearer <token>`, validasi signature & expiry, inject `user_id`/`role` ke Gin context |
| Middleware RBAC | Cek permission spesifik per-endpoint, contoh: `router.DELETE("/users/:id", middleware.RequirePermission("user.delete"), ...)` |
| Logout | Blacklist token di Redis (atau hapus refresh token dari DB) |

**Endpoint terkait:**
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET  /api/v1/auth/me`

---

## 6. Modul Keamanan (Security)

| Item | Spesifikasi |
|---|---|
| CORS | Whitelist domain via env `ALLOWED_ORIGINS` (contoh: `http://localhost`, `https://ponpes.ferxcode.my.id`), menolak origin lain secara default (bukan `*`) |
| Rate Limiting | Default: 100 request/menit per IP untuk endpoint publik, lebih ketat untuk `/auth/login` (misal 5x/menit) untuk mencegah brute force. In-memory (`golang.org/x/time/rate`) untuk single instance, atau Redis untuk multi-instance |
| Security Headers | `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `X-XSS-Protection`, `Strict-Transport-Security`, `Content-Security-Policy` dasar. Rekomendasi pakai library `secure` (unrolled/secure) |
| Input Validation | Gin binding + `validator` tag (`required`, `email`, `min`, `max`), sanitasi HTML pada field longtext (article content) menggunakan `bluemonday` untuk mencegah stored XSS |
| SQL Injection | Otomatis ter-mitigasi karena GORM pakai parameterized query — tetap wajib hindari raw query string concatenation |
| CSRF | Tidak wajib untuk API stateless berbasis JWT (bukan cookie session), tapi jika refresh token disimpan di cookie → wajib `HttpOnly`, `Secure`, `SameSite=Strict` |

---

## 7. Struktur Arsitektur Berbasis Modul

Sudah dijabarkan di Bab 3. Poin tambahan:

- **Config Management:** Viper membaca `.env` (development) dan bisa fallback ke environment variable asli (production, misal di Docker/Cloud Run). Struct `Config` di-inject ke semua layer via dependency injection sederhana (constructor injection), bukan global variable, agar testable.
- **Dependency Injection:** Gunakan pola manual constructor injection (`NewUserService(repo Repository) *Service`) di `main.go`. Untuk proyek kecil ini cukup, tidak perlu wire/fx kecuali tim ingin standar lebih ketat.
- **Routing:** `router.go` sebagai composition root, memanggil `RegisterPublicRoutes()` dan `RegisterAdminRoutes()` yang masing-masing mount middleware berbeda (admin routes wajib lewat `AuthMiddleware` + `RBACMiddleware`).

---

## 8. Penanganan Error & Logging

### 8.1 Standardized JSON Response

```json
// Success
{
  "status": "success",
  "message": "Article created successfully",
  "data": { "id": 1, "title": "..." }
}

// Error
{
  "status": "error",
  "message": "Validation failed",
  "errors": {
    "email": "must be a valid email address"
  }
}
```

Diimplementasikan via helper `response.Success(c, data, message)` dan `response.Error(c, statusCode, message, errors)` agar konsisten di semua handler.

### 8.2 Global Error Handler
- Middleware `recovery_middleware.go` menangkap `panic`, log stack trace, dan mengembalikan response 500 standar (tidak expose stack trace ke client di production).
- Custom `AppError` type (dengan `Code`, `Message`, `HTTPStatus`) agar service layer bisa melempar error terstruktur yang diterjemahkan otomatis oleh handler.

### 8.3 Structured Logging
- Gunakan `zerolog`, log dalam format JSON: `timestamp`, `method`, `path`, `status_code`, `latency_ms`, `client_ip`, `request_id`.
- **Request ID Middleware**: setiap request diberi UUID unik untuk tracing (memudahkan korelasi log saat debugging).
- Log level dikontrol via env (`debug`/`info`/`warn`/`error`), default `info` di production.

---

## 9. Manajemen File & Upload

| Aspek | Spesifikasi |
|---|---|
| Metode | `multipart/form-data`, field `file` |
| Validasi ekstensi | Whitelist: `.jpg`, `.jpeg`, `.png`, `.webp` (untuk gambar), cek MIME type aktual (bukan hanya ekstensi) via `http.DetectContentType` untuk mencegah file disguise |
| Batas ukuran | Default 2MB (configurable via env `MAX_UPLOAD_SIZE`), pakai `MaxMultipartMemory` di Gin |
| Penamaan file | UUID + timestamp untuk hindari collision, simpan nama asli di kolom `file_name` tabel `media_assets` |
| Static serving | `router.Static("/uploads", "./storage/uploads")`, dengan opsi tambahan header cache-control untuk aset statis |
| Struktur folder upload | `storage/uploads/{module}/{yyyy}/{mm}/filename.ext` agar rapi dan mudah di-backup per periode |

---

## 10. DevOps & Infrastruktur

### 10.1 Dockerfile (Multi-stage, dengan target `dev` dan `prod`)

MySQL **dan** Redis **tidak** disediakan di dalam boilerplate ini — aplikasi akan connect ke **shared MySQL container** dan **shared Redis container** yang sudah berjalan terpisah (dikelola di luar proyek ini, dipakai bersama beberapa proyek company profile). Konfigurasi koneksi (host, port, user, password, nama DB/prefix) cukup diarahkan lewat env ke container/service tersebut, dan pastikan network Docker-nya terhubung (lihat `docker-compose.yml` di bawah, bagian `networks: external`).

Dockerfile dibuat dengan **multi-stage** yang memisahkan target `dev` (pakai [`air`](https://github.com/air-verse/air) untuk hot reload) dan `prod` (binary final yang ramping):

```dockerfile
# syntax=docker/dockerfile:1

# ---------- Base ----------
FROM golang:1.22-alpine AS base
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download

# ---------- Dev (hot reload) ----------
FROM base AS dev
RUN go install github.com/air-verse/air@latest
COPY . .
EXPOSE 8080
# air akan watch perubahan source code & auto-rebuild+restart
CMD ["air", "-c", ".air.toml"]

# ---------- Build (compile binary untuk production) ----------
FROM base AS build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/api

# ---------- Prod (image final, ramping) ----------
FROM alpine:3.19 AS prod
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /app/server .
COPY --from=build /app/storage ./storage
EXPOSE 8080
ENTRYPOINT ["./server"]
```

- Stage `dev`: source code di-*mount* sebagai volume (lihat compose di bawah), sehingga `air` mendeteksi perubahan file `.go` secara langsung tanpa perlu rebuild image, cocok untuk development lokal.
- Stage `prod`: hasil akhir hanya binary + alpine base (biasanya <20MB), tanpa toolchain Go maupun source code mentah, sesuai standar image production yang ramping & aman.

### 10.2 docker-compose.yml — profile `dev` & `prod`, tanpa container MySQL sendiri

Compose hidup di **root monorepo** (bukan di dalam `backend/`). Satu file `docker-compose.yml` menjalankan mode `dev` (hot reload) atau `prod` (binary optimal + Nuxt Nitro) tanpa duplikasi file. MySQL dan Redis diasumsikan sudah berjalan sebagai **shared container** di luar stack ini — dihubungkan lewat Docker **external network** `app-bridge`.

Service aktual:

| Profile | Service | Peran |
|---|---|---|
| `dev` | `api-dev`, `web-dev`, `nginx` | Go (air), Nuxt (`nuxt dev`), nginx `:80` |
| `prod` | `api`, `web`, `nginx-prod` | Go binary, Nitro `:3000`, nginx `:8080` |

Nginx memproksi `/` → Nuxt, `/api/` → Go, `/uploads/` → Go. Env gabungan ada di root `.env` (dari `.env.example`), bukan `.env.development` / `.env.production` terpisah.

```yaml
# cuplikan — lihat docker-compose.yml di root repo
services:
  api-dev:
    profiles: ["dev"]
    build:
      context: ./backend
      target: dev
    env_file: ./.env
    networks:
      - shared-net

  api:
    profiles: ["prod"]
    build:
      context: ./backend
      target: prod
    env_file: ./.env
    networks:
      - shared-net

networks:
  shared-net:
    external: true
    name: app-bridge
```

**Cara pakai:**

```bash
# Mode development (hot reload via air + nuxt dev)
make dev
# docker compose --profile dev up --build

# Mode production
make prod
# docker compose --profile prod up --build -d
```

**Catatan penting terkait shared MySQL & shared Redis:**
- `DB_HOST` di `.env` diisi dengan **service name / container name MySQL bersama** tersebut (bukan `mysql` lokal lagi), contoh: `DB_HOST=shared-mysql`.
- `REDIS_HOST` (dipakai untuk rate limiting terdistribusi & blacklist token, lihat Bab 5–6) juga diarahkan ke **service name/container name Redis bersama**, contoh: `REDIS_HOST=shared-redis`. Boilerplate ini tidak lagi menyediakan container Redis sendiri.
- Pastikan network `app-bridge` sudah dibuat lebih dulu (`make network` atau `docker network create app-bridge`) dan container MySQL **serta** Redis bersama sudah join ke network yang sama, agar service `api-dev`/`api` bisa resolve hostname keduanya.
- Karena database sekarang **shared**, wajib pastikan **nama database unik per proyek** (`DB_NAME`) untuk menghindari bentrok skema antar proyek yang memakai MySQL bersama yang sama.
- Karena Redis juga **shared**, wajib pastikan **key prefix unik per proyek** (misal `REDIS_PREFIX=cp_projectname_`) untuk key rate-limit, cache, atau token blacklist — agar tidak bentrok/tertimpa data proyek lain yang memakai Redis bersama yang sama.
- Auto-migration & seeder (lihat 10.3–10.4) tetap berjalan seperti biasa saat container app start — hanya target koneksi MySQL & Redis yang berubah dari container lokal menjadi shared container.

### 10.3 Auto Migration
- Dijalankan otomatis saat aplikasi start (`db.AutoMigrate(&User{}, &Role{}, ...)`), dengan urutan yang memperhatikan foreign key dependency.
- Tambahkan **wait-for-db** logic (retry connect ke MySQL beberapa kali) karena container MySQL sering start lebih lambat dari app container.

### 10.4 Database Seeder
Berjalan sekali (idempotent — cek `if not exists` sebelum insert) saat pertama kali deploy:
- Insert 3 role default: `SuperAdmin`, `Admin`, `Editor`.
- Insert permission dasar (`user.*`, `article.*`, `service.*`, `setting.*`, dst).
- Insert 1 akun SuperAdmin default (email & password dari env, **wajib diganti setelah first login**).
- Insert `web_settings` default (nama web, logo placeholder, dsb).

---

## 11. Utilitas Tambahan

### 11.1 Pagination Standar
```json
{
  "status": "success",
  "data": [ ... ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total_data": 42,
    "total_pages": 5
  }
}
```
Helper generik `pagination.Paginate(db, page, perPage)` dipakai di semua endpoint list (articles, contact_messages, users, dst).

### 11.2 API Documentation
- Swagger via `swaggo/swag`, comment annotation di tiap handler.
- Endpoint docs: `GET /swagger/index.html` (hanya aktif di environment non-production, atau dilindungi basic auth jika tetap ingin diakses di production).

---

## 12. Modul Tambahan yang Direkomendasikan (Standar Industri, di luar daftar awal)

Berikut modul yang **belum disebutkan** di requirement awal tapi umum ada di boilerplate backend production-grade:

| Modul | Alasan |
|---|---|
| **Health Check Endpoint** (`GET /healthz`, `GET /readyz`) | Dibutuhkan oleh load balancer/orchestrator (atau sekadar monitoring uptime) untuk cek app & koneksi DB masih hidup |
| **Graceful Shutdown** | Menangani `SIGTERM`/`SIGINT` agar request yang sedang berjalan selesai dulu sebelum container mati (penting saat `docker stop` atau redeploy) |
| **Environment-based Config Validation** | Validasi saat startup bahwa semua env variable wajib (DB_HOST, JWT_SECRET, dsb) terisi — fail fast jika ada yang kosong, daripada error di tengah jalan |
| **Email Service (SMTP)** | Notifikasi otomatis saat ada `contact_messages` baru masuk ke email admin, juga berguna untuk fitur forgot-password nantinya |
| **Forgot/Reset Password Flow** | Umum dibutuhkan meski tidak disebut eksplisit — token reset via email, expiry singkat |
| **Audit Log / Activity Log** | Tabel `activity_logs` (siapa melakukan apa, kapan) — penting untuk akuntabilitas di CMS multi-admin |
| **Database Connection Pooling** | Konfigurasi eksplisit `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime` pada GORM/sql.DB, agar tidak exhaust koneksi MySQL |
| **Unit Test & Integration Test Structure** | Folder `*_test.go` per modul, mock repository via interface, minimal test untuk service layer (auth, RBAC) |
| **CI Pipeline (GitHub Actions)** | Job: lint (`golangci-lint`) → test → build Docker image, jalan otomatis di setiap PR |
| **API Versioning** | Prefix `/api/v1/...` sejak awal, agar breaking change ke depan tidak merusak client existing |
| **Environment Separation** | Satu `.env` di root monorepo (dari `.env.example`); bedakan `APP_ENV` / CORS / log level per environment |
| **Backup Strategy (dokumentasi, bukan kode)** | Cron `mysqldump` terjadwal + rotasi backup volume upload, didokumentasikan di README meski implementasi di luar scope kode boilerplate |

---

## 13. Non-Functional Requirements

| Aspek | Target |
|---|---|
| Response time API (non-upload) | < 200ms rata-rata pada beban normal |
| Ukuran image Docker final | < 30MB |
| Startup time container | < 5 detik (di luar waktu tunggu DB ready) |
| Test coverage minimum | 60% untuk service layer auth & RBAC |
| Kompatibilitas | MySQL 8.x, Go 1.22+, Docker Compose v2 |

---

## 14. Daftar Endpoint (Ringkasan)

| Method | Endpoint | Akses |
|---|---|---|
| POST | `/api/v1/auth/login` | Public |
| POST | `/api/v1/auth/refresh` | Public (dengan refresh token) |
| POST | `/api/v1/auth/logout` | Authenticated |
| GET | `/api/v1/auth/me` | Authenticated |
| GET | `/api/v1/public/articles` | Public |
| GET | `/api/v1/public/articles/:slug` | Public |
| GET | `/api/v1/public/services` | Public |
| GET | `/api/v1/public/settings` | Public |
| POST | `/api/v1/public/contact` | Public (rate-limited ketat) |
| CRUD | `/api/v1/admin/users` | SuperAdmin/Admin (RBAC) |
| CRUD | `/api/v1/admin/roles` | SuperAdmin |
| CRUD | `/api/v1/admin/articles` | Admin/Editor |
| CRUD | `/api/v1/admin/services` | Admin/Editor |
| GET/PATCH | `/api/v1/admin/contact-messages` | Admin |
| GET/PUT | `/api/v1/admin/settings` | SuperAdmin/Admin |
| POST | `/api/v1/admin/media/upload` | Admin/Editor |
| GET | `/healthz`, `/readyz` | Public (infra) |
| GET | `/swagger/*` | Public/dilindungi (non-prod) |

---

## 15. Acceptance Criteria

Boilerplate dianggap selesai jika:

1. `docker compose --profile dev up --build` menjalankan app dengan hot reload (perubahan file `.go` langsung ter-*restart* tanpa rebuild image), dan `docker compose --profile prod up --build -d` menjalankan image production yang ramping — keduanya berhasil connect ke shared MySQL container dan shared Redis container (via external network) tanpa perlu container MySQL/Redis sendiri, auto-migrate & seeder tetap berjalan otomatis.
2. Login SuperAdmin default berhasil, JWT diterbitkan, endpoint admin ter-protect dan menolak role yang tidak sesuai (RBAC teruji).
3. Endpoint publik (articles, services, settings, contact) bisa diakses tanpa token dan CORS hanya mengizinkan domain yang di-whitelist.
4. Upload gambar tervalidasi (tipe & ukuran), tersimpan di `storage/uploads`, tercatat di `media_assets`, dan bisa diakses via static URL.
5. Semua response API mengikuti format standar (`status`, `message`, `data`/`errors`).
6. Swagger docs bisa diakses dan mencakup seluruh endpoint di atas.
7. Rate limiting terbukti memblokir request berlebih pada endpoint login & contact form.
8. README mendokumentasikan cara setup, struktur env, dan cara menambah modul baru mengikuti pola Handler→Service→Repository yang sudah ada.

---

## 16. Saran Urutan Pengerjaan (Milestone)

1. **Fase 1 – Fondasi:** Config, koneksi DB, struktur folder, standardized response, logger, error handler, graceful shutdown, health check.
2. **Fase 2 – Auth & RBAC:** Model users/roles/permissions, JWT, bcrypt, middleware auth & RBAC, seeder SuperAdmin.
3. **Fase 3 – Modul Konten:** web_settings, articles, services, contact_messages, pagination.
4. **Fase 4 – Media & Security Hardening:** Upload module, CORS, rate limiting, security headers, input sanitization.
5. **Fase 5 – DevOps & Docs:** Dockerfile multi-stage, docker-compose, Swagger, CI pipeline, README, unit test dasar.

---

*Dokumen ini bersifat living document — dapat disesuaikan sesuai kebutuhan spesifik tiap proyek company profile yang menggunakan boilerplate ini.*

# PRD: Boilerplate Frontend Nuxt 3 — Landing Page & Company Profile dengan Admin Panel

| Field | Detail |
|---|---|
| **Nama Proyek** | Nuxt 3 Boilerplate — Company Profile + Admin Panel |
| **Versi Dokumen** | 2.0 |
| **Target Deploy** | Docker + nginx di VPS (satu stack dengan backend Go) |
| **Backend** | Go (Gin) — konsumsi via REST API |
| **Status** | Draft untuk Review |

---

## 1. Ringkasan Eksekutif

Dokumen ini mendefinisikan kebutuhan produk untuk membangun **boilerplate frontend** berbasis Nuxt 3 yang dapat digunakan berulang kali sebagai fondasi proyek company profile / landing page dengan panel admin. Boilerplate ini menjadi starter kit standar tim agar setiap proyek baru tidak perlu membangun ulang arsitektur dasar (auth, routing, data fetching, komponen UI) dari nol.

Boilerplate mencakup dua sisi aplikasi dalam satu codebase Nuxt:
1. **Sisi Publik** — landing page/company profile (SSR disarankan untuk SEO; kondisi awal template adalah SPA).
2. **Sisi Admin** — panel manajemen konten yang aman, dilindungi autentikasi berbasis JWT dan RBAC, mengonsumsi backend Go Gin.

**Kondisi awal (versi ini):** UI admin memakai template Spike (Vuetify). `ssr: false`. Integrasi API, auth, dan halaman publik company profile belum diimplementasikan — lihat non-goals / future consideration.

---

## 2. Latar Belakang & Masalah yang Diselesaikan

- Setiap proyek company profile baru membutuhkan pola yang sama: landing page publik + admin panel untuk mengelola konten (artikel, pesan masuk, pengaturan web).
- Tanpa boilerplate, tim berulang kali menulis ulang setup autentikasi, proteksi rute, data table, form validation, dan SEO metadata — memakan waktu dan rawan inkonsistensi.
- Dibutuhkan starter kit yang **production-ready**, mengikuti standar industri, ringan, dan mudah di-maintain oleh developer baru.
- Frontend dan backend di-deploy bersama di VPS via Docker Compose; nginx adalah satu-satunya entry publik.

---

## 3. Tujuan Produk (Goals)

1. Mempercepat *time-to-first-deploy* proyek baru dari hitungan minggu menjadi hitungan hari.
2. Menstandarkan arsitektur, konvensi kode, dan pola keamanan di seluruh proyek company profile.
3. Memberikan skor SEO & Core Web Vitals yang baik setelah SSR diaktifkan untuk halaman publik.
4. Menyediakan admin panel modular yang mudah diperluas (tambah modul baru tanpa merombak struktur inti).
5. Kompatibel penuh dengan deployment Docker + nginx (`make dev` / `make prod`).

### Non-Goals (Di Luar Lingkup)

- Tidak mencakup pengembangan backend Go Gin (diasumsikan sudah tersedia/berjalan paralel).
- Tidak termasuk desain visual final (branding/UI kit spesifik klien) — boilerplate menyediakan *design system* generik berbasis Vuetify yang bisa di-theming.
- Tidak mencakup fitur e-commerce, payment gateway, atau multi-tenant.
- Tidak mencakup aktivasi SSR di putaran fondasi Docker (tetap `ssr: false` sampai halaman publik dibangun).

---

## 4. Target Pengguna

| Persona | Kebutuhan |
|---|---|
| **Developer internal (tim FE)** | Starter kit yang jelas strukturnya, terdokumentasi, mudah di-extend per proyek. |
| **Admin/Editor konten (end user aplikasi)** | Panel yang mudah dipakai untuk mengelola artikel, pesan, dan pengaturan situs. |
| **Pengunjung situs publik** | Landing page cepat, mobile-friendly, mudah ditemukan di mesin pencari. |
| **Project Manager/Klien** | Kepercayaan bahwa hasil akhir aman, cepat, dan sesuai standar industri. |

---

## 5. Lingkup Fitur (Scope)

### 5.1 In-Scope

- Landing page publik dengan SEO dinamis (setelah SSR diaktifkan)
- Autentikasi admin (login, refresh token, logout, proteksi rute)
- RBAC (role-based access control) di level UI
- Dashboard admin dengan statistik ringkas
- Manajemen pengguna & peran
- Manajemen artikel/konten dengan rich text editor
- Manajemen pengaturan web (key-value settings)
- Kotak masuk pesan kontak
- Media library terintegrasi API upload Go
- Konfigurasi siap deploy Docker (env vars, nginx `/`, `/api/`, `/uploads/`)

### 5.2 Out-of-Scope (versi ini)

- Multi-bahasa (i18n) — dicatat sebagai *future consideration*
- Notifikasi real-time (WebSocket/push notification)
- Analytics dashboard custom (cukup integrasi Google Analytics dasar)
- Deploy Vercel / hosting Node terpisah dari stack Docker

---

## 6. Arsitektur & Tech Stack

| Layer | Pilihan Teknologi | Justifikasi |
|---|---|---|
| Framework | **Nuxt 3** | File-based routing, Nitro server, `runtimeConfig` untuk env server/public; standar industri Vue SSR/SPA. |
| UI | **Vuetify 3** (template Spike sebagai titik awal) | Komponen admin siap pakai, aksesibilitas Material, mudah di-theming. |
| Ikon | **MDI / Tabler Icons / Iconify** | Sudah termasuk di template Spike. |
| State Global | **Pinia** (saat modul admin diimplementasikan) | Standar resmi Vue; minim boilerplate. |
| Data Fetching | **`$fetch` / ofetch** + opsi Pinia/useAsyncData | Native Nuxt; caching via `useAsyncData` / `useFetch`. |
| Form Management | **VeeValidate** atau native + Zod | Validasi client-side yang selaras dengan validasi backend Go. |
| Validasi Skema | **Zod** | Validasi client-side yang selaras dengan validasi backend Go. |
| Rich Text Editor | **TipTap** (pilihan utama) | Editor WYSIWYG standar industri; dimuat client-only. |
| HTTP Client | **`$fetch` dengan interceptor** (plugin Nuxt) | Auto-inject Bearer token & auto-refresh token saat 401. |
| Autentikasi Sesi | **Implementasi kustom** | JWT dari backend Go. Token akses di memory / cookie HttpOnly via Nitro jika diperlukan. **Jangan** memakai NextAuth/Auth.js. |
| Deploy | **Docker + nginx** | Satu origin: `/` → Nuxt, `/api/` → Go, `/uploads/` → Go. |

Kondisi awal: `ssr: false` (SPA). **Future consideration:** aktifkan SSR untuk rute publik (`/`, `/artikel`, `/kontak`, dsb.) dan biarkan rute `/admin` tetap client-heavy.

---

## 7. Kebutuhan Fungsional per Modul

### 7.1 Autentikasi, Sesi, & Middleware

- **FR-1**: Sistem login admin menggunakan JWT. Kredensial tidak boleh disimpan di `localStorage` jika bisa dihindari; utamakan cookie HttpOnly via Nitro server route, atau memory + refresh token rotation. Autentikasi **wajib** kustom — **bukan** NextAuth.js/Auth.js.
- **FR-2**: Nuxt route middleware memblokir akses ke seluruh rute `/admin/*` bagi user yang belum login, dan redirect ke `/login` (atau `/auth/login`).
- **FR-3**: Plugin HTTP secara otomatis menyisipkan Bearer token di setiap request ke API Go (same-origin `/api/v1` di belakang nginx).
- **FR-4**: Ketika response API mengembalikan 401 Unauthorized, sistem otomatis memanggil endpoint `/api/v1/auth/refresh`; jika gagal, user di-redirect ke halaman login.
- **FR-5**: Komponen wrapper / composable `usePermission("article.create")` untuk menyembunyikan/menampilkan elemen UI berdasarkan permission dari payload JWT.

### 7.2 Penanganan Data & Formulir

- **FR-6**: Semua halaman list data admin (artikel, user, pesan) menggunakan Data Table dengan sorting, pagination, dan filtering.
- **FR-7**: Semua formulir input admin memakai validasi skema (Zod atau VeeValidate) sebelum submit ke API.
- **FR-8**: Sistem menampilkan pesan error validasi per-field secara real-time (inline error messages).
- **FR-8a**: Komponen Rich Text Editor (TipTap) **wajib** dimuat client-only (`ClientOnly` + `defineAsyncComponent`) dan hanya di-load pada halaman yang membutuhkannya. Editor tidak boleh ikut termuat pada bundle halaman lain.

### 7.3 Landing Page & SEO

- **FR-9**: Setiap halaman publik memiliki metadata dinamis (`useSeoMeta` / `useHead`), diambil dari data `web_settings`/`articles` backend. Pengambilan data memakai `useAsyncData` dengan cache/TTL, bukan fetch tanpa cache pada setiap request. Nilai revalidasi disarankan disesuaikan per jenis data.
- **FR-10**: Gambar memakai `<NuxtImg>` (atau fallback `<img>` dengan placeholder) dan **wajib** punya mekanisme *fallback image* apabila URL dari backend error/404.
- **FR-11**: Font di-self-host atau dimuat dengan `font-display: swap` untuk mencegah *layout shift* (CLS).
- **FR-12**: Sitemap dan robots digenerate (modul `@nuxtjs/sitemap` / route Nitro) berdasarkan data artikel & halaman terbaru — berlaku setelah SSR/static generation diaktifkan.

### 7.4 Admin Panel — Modul Fungsional

- **FR-13 (Dashboard)**: Menampilkan ringkasan statistik (jumlah pesan masuk, total artikel, dll.) dari API.
- **FR-14 (Manajemen Pengguna & Peran)**: CRUD user admin/editor, penetapan role, tabel dengan sorting & pagination.
- **FR-15 (Manajemen Konten/Artikel)**: CRUD artikel dengan rich text editor, upload thumbnail, pengaturan status Draft/Terbit.
- **FR-16 (Web Settings)**: Form dinamis key-value untuk mengubah logo, nama situs, footer, dan tautan media sosial.
- **FR-17 (Kotak Masuk)**: Daftar pesan dari form kontak landing page, dengan aksi "Tandai Sudah Dibaca".
- **FR-18 (Media Library)**: Modal/panel pemilih media yang terhubung ke API upload Go Gin, mendukung pilih gambar existing atau upload baru.
- **FR-19 (Manajemen Banner/Slider)**: Antarmuka khusus bagi admin untuk mengelola Hero Banner/Slider di beranda. Mencakup tambah/hapus item, unggah/pilih gambar, tautan & caption, serta drag-and-drop reorder.

---

## 8. Kebutuhan Non-Fungsional

| Kategori | Kebutuhan |
|---|---|
| **Performa** | Core Web Vitals hijau (LCP < 2.5s, CLS < 0.1, INP < 200ms) pada landing page (setelah SSR). |
| **SEO** | Skor Lighthouse SEO ≥ 90; metadata dinamis lengkap per halaman (setelah SSR). |
| **Keamanan** | Token JWT tidak disimpan di `localStorage` jika ada alternatif HttpOnly; validasi input di client & server. |
| **Aksesibilitas** | Komponen UI mengikuti standar WCAG dasar (mewarisi aksesibilitas Vuetify). |
| **Skalabilitas Kode** | Struktur folder modular Nuxt (`pages`, `components`, `composables`, `middleware`). |
| **Maintainability** | ESLint terkonfigurasi; konsisten di seluruh tim. |
| **Kompatibilitas Deploy** | `make dev` / `make prod`; environment variables terdokumentasi di `.env.example`. |
| **Bundle Size** | Editor & chart di-load on-demand; tidak ada dependency raksasa tak terpakai. |

---

## 9. Struktur Halaman (Sitemap Aplikasi)

**Publik:**
- `/` — Beranda
- `/tentang-kami`
- `/layanan` atau `/services`
- `/artikel` (blog listing) & `/artikel/[slug]` (detail)
- `/kontak`

**Admin (dilindungi middleware):**
- `/login` atau `/auth/login`
- `/admin` — Dashboard
- `/admin/users` — Manajemen Pengguna
- `/admin/articles` — Manajemen Artikel
- `/admin/banners` — Manajemen Banner/Slider
- `/admin/settings` — Web Settings
- `/admin/messages` — Kotak Masuk
- `/admin/media` — Media Library

---

## 10. Integrasi Backend (API Go Gin)

- Seluruh komunikasi data admin & publik menggunakan REST API dari backend Go Gin.
- Di Docker, nginx memproksi `/api/` ke service Go sehingga browser memakai **same-origin**. `NUXT_PUBLIC_API_URL` default ke origin publik (contoh `http://localhost`).
- `API_URL` (server-only) diisi hostname internal Compose (`http://api-dev:8080` / `http://api:8080`) untuk Nitro server routes.
- Skema validasi Zod di frontend diselaraskan secara manual dengan skema validasi Go di backend. Idealnya generate dari OpenAPI.
- Endpoint kunci: `/api/v1/auth/login`, `/api/v1/auth/refresh`, `/api/v1/public/articles`, `/api/v1/admin/users`, `/api/v1/admin/settings`, `/api/v1/public/contact`, `/api/v1/admin/media/upload`.

---

## 11. Operasional & Standar Deployment (Docker)

- **Environment Variables**: `.env` di root repo (dari `.env.example`). Compose meng-inject ke semua service.
- **Images**: `frontend/Dockerfile.dev` (`nuxt dev`), `frontend/Dockerfile` (Nitro `node .output/server/index.mjs`, non-root).
- **Linting**: job CI `npm ci` + `npm run build` pada setiap PR yang menyentuh `frontend/**`.
- **Compose validation**: `docker compose --profile dev|prod config` di CI.
- **Healthcheck**: `web` / `web-dev` mengecek `GET /` pada port 3000.

---

## 12. Milestone & Rencana Rilis

| Fase | Cakupan | Estimasi |
|---|---|---|
| **Fase 0** | Docker, env `NUXT_PUBLIC_*`, nginx `/api/`, dokumentasi (putaran ini) | 1–2 hari |
| **Fase 1** | Struktur folder, ESLint, composables API, `runtimeConfig` terpakai | 2–3 hari |
| **Fase 2** | Autentikasi, middleware proteksi rute, interceptor, RBAC wrapper | 3–4 hari |
| **Fase 3** | Modul admin: Dashboard, Manajemen Pengguna, Web Settings | 4–5 hari |
| **Fase 4** | Modul admin: Manajemen Artikel (rich text editor, media library), Kotak Masuk | 5–6 hari |
| **Fase 5** | Landing page publik + aktifkan SSR + SEO (`useSeoMeta`, sitemap) | 3–4 hari |
| **Fase 6** | QA, audit performa (Lighthouse), dokumentasi boilerplate | 2–3 hari |

*Total estimasi setelah Fase 0: ± 3–4 minggu kerja untuk satu developer.*

---

## 13. Metrik Keberhasilan (Success Metrics)

- Waktu setup proyek baru dari boilerplate ini ≤ 1 hari kerja.
- Skor Lighthouse (setelah SSR): Performance ≥ 90, SEO ≥ 90, Accessibility ≥ 90 pada landing page.
- Zero critical vulnerability pada audit dependency (`npm audit`).
- Dokumentasi boilerplate cukup lengkap sehingga developer baru bisa onboarding tanpa bantuan langsung dalam < 2 jam.
- `make dev` dan `make prod` berhasil tanpa mengubah Dockerfile.

---

## 14. Risiko & Mitigasi

| Risiko | Dampak | Mitigasi |
|---|---|---|
| Ketidaksesuaian skema validasi Zod (FE) vs Go (BE) | Bug data tidak konsisten | Dokumentasikan skema bersama, idealnya generate dari OpenAPI |
| Token refresh gagal menyebabkan user ter-logout mendadak | UX admin terganggu | Retry & fallback redirect yang jelas ke halaman login |
| Bundle size membengkak karena rich text editor besar | Performa admin menurun | `ClientOnly` + `defineAsyncComponent` (FR-8a) |
| `ssr: false` merugikan SEO halaman publik | Landing page sulit diindeks | Aktifkan SSR pada Fase 5; biarkan admin tetap client-heavy |
| URL gambar dari backend rusak/404 | Layout rusak | Fallback image wajib (FR-10) |

---

## 15. Lampiran — Contoh Struktur Folder (Disarankan)

```
frontend/
├── pages/
│   ├── index.vue
│   ├── tentang-kami.vue
│   ├── layanan.vue
│   ├── artikel/
│   ├── kontak.vue
│   ├── auth/
│   └── admin/
├── components/
│   └── shared/
├── composables/
│   ├── useApi.ts
│   └── useAuth.ts
├── middleware/
│   └── auth.ts
├── layouts/
├── plugins/
├── server/            # Nitro routes (opsional BFF)
├── nuxt.config.ts
├── Dockerfile
├── Dockerfile.dev
└── types/
```

---

*Dokumen ini merupakan draft PRD yang siap direview oleh tim engineering dan stakeholder sebelum masuk ke fase development.*

# Go-Nuxt Company Profile Boilerplate (Monorepo)

Boilerplate ini adalah *starter kit* berarsitektur **Monorepo** untuk membangun website *Company Profile* atau *Landing Page* dinamis. Repositori ini memuat dua lingkungan yang terpisah secara eksekusi namun dikelola dalam satu basis kode:

1. **Frontend:** Nuxt 3 + Vuetify untuk antarmuka publik dan panel admin.
2. **Backend:** Go (Gin) REST API untuk manajemen konten dan basis data.

Keduanya di-deploy sebagai container Docker di server yang sama. Nginx adalah satu-satunya entry publik.

---

## Struktur Direktori Utama

Repositori ini secara tegas memisahkan area *frontend* dan *backend*. Tolong **jangan** mencampur *dependency* atau konfigurasi antar-folder.

```text
nuxt-go/
├── frontend/                  # Nuxt 3 + Vuetify (Docker)
│   ├── pages/                 # File-based routing
│   ├── components/
│   ├── layouts/
│   ├── plugins/
│   ├── public/                # Aset statis
│   ├── server/                # Nitro server (siap untuk route/BFF)
│   ├── nuxt.config.ts
│   ├── package.json
│   ├── Dockerfile
│   └── Dockerfile.dev
│
├── backend/                   # Go Gin, GORM, MySQL
│   ├── cmd/api/               # Entry point Go (main.go)
│   ├── internal/              # Core logic (Domain, Handler, Service, Repo)
│   ├── storage/uploads/       # Direktori lokal untuk file gambar (Docker Volume)
│   ├── go.mod
│   └── Dockerfile             # Multi-stage build (target dev / prod)
│
├── deploy/nginx/              # Snippet nginx (dev / prod / global TLS)
├── docker-compose.yml         # Profile dev + prod (api, web, nginx)
├── .env.example               # Env gabungan backend + frontend
├── Makefile
├── .gitignore
└── README.md
```

Env hidup di root (`.env` dari `.env.example`). Compose meng-inject file yang sama ke semua service dan men-override `API_URL` ke hostname internal (`http://api-dev:8080` / `http://api:8080`).

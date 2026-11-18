# Golang API Boilerplate

CRUD Posts & Categories dengan Gin + `database/sql` + PostgreSQL + Redis,
mencakup autentikasi JWT (register, login, refresh token, logout,
forgot/reset password) dan verifikasi email via OTP.

## Struktur folder

```
cmd/server/main.go                     # entry point - load config, connect DB & Redis, wiring, HTTP server + graceful shutdown
internal/
  config/config.go                     # load .env + env (DB_*/POSTGRES_*, REDIS_*, JWT_*, OTP_*), validasi fail-fast
  database/database.go                 # koneksi database/sql PostgreSQL + migrasi manual saat startup
  database/redis.go                    # koneksi Redis (cache / session / rate limit)
  model/                               # entitas + kontrak repository + error domain
    post.go, category.go               #   Post, Category (soft delete)
    user.go, role.go                   #   User, Role
    otp.go, refresh_token.go           #   OTP (hash), RefreshToken
  repositories/                        # adapter database/sql untuk tiap kontrak model
    post_repository.go                 #   model.PostRepository
    category_repository.go             #   model.CategoryRepository
    user_repository.go                 #   model.UserRepository
    role_repository.go                 #   model.RoleRepository
    otp_repository.go                  #   model.OTPRepository
    refresh_token_repository.go        #   model.RefreshTokenRepository
  services/                            # usecase / business logic (hanya impor model)
    post_service.go
    category_service.go
    auth_service.go                    #   register, login, refresh, logout, OTP, reset password
  dto/                                 # DTO request/response (binding + json tag)
    post_dto.go, category_dto.go
    auth_dto.go
  handlers/                            # interface adapter HTTP (binding, validasi, response)
    post_handler.go
    category_handler.go
    auth_handler.go
    health_handler.go                  #   health check (ping DB & Redis)
  middleware/
    auth.go                            #   validasi bearer access token -> claims ke gin.Context
    ratelimit.go                       #   rate limit per IP (golang.org/x/time/rate)
    security.go                        #   HTTP security headers (CSP, HSTS, X-Frame-Options, dll.)
  routes/router.go                     # registrasi route + pemasangan middleware
pkg/
  response/response.go                 # envelope response standar
  logger/logger.go                     # wrapper logging
  token/token.go                       # generate/parse JWT access token + hash refresh token
  validator/validator.go               # format error validasi (murni, tanpa gin)
  validator/auth_validator.go          # validator khusus auth (eqfield untuk confirm_password)
  mailer/mailer.go                     # kirim OTP via Brevo HTTP API (interface Mailer)
docs/                                  # hasil generate swag (docs.go, swagger.json, swagger.yaml)
Dockerfile                             # multi-stage build (builder -> alpine runtime, non-root)
docker-compose.yml                     # app + postgres + redis
.dockerignore                          # kecualikan .env, binary, .git dari build context
.env.example                           # contoh konfigurasi
```

Alur dependency:
`main → routes → handlers → services → model ← repositories (database/sql)`.
Wiring repo -> service -> handler dilakukan langsung di `main.go`.
Handler tidak menyentuh database; service hanya mengimpor `model`;
repository mengimplementasikan kontrak `model.*Repository` dan menerjemahkan
`sql.ErrNoRows` menjadi error domain (mis. `model.ErrPostNotFound`,
`model.ErrCategoryNotFound`). Pemetaan bentuk data API dilakukan oleh DTO.

Catatan per fitur:

- **Category** memakai soft delete (kolom `deleted_at`).
- **Auth** memakai access token JWT berumur pendek + refresh token yang
  disimpan sebagai hash di database, sehingga bisa di-revoke saat logout.
- **OTP** juga disimpan sebagai hash, punya masa berlaku (`OTP_EXPIRE_MINUTES`)
  dan batas percobaan (`OTP_MAX_ATTEMPTS`).
- **Email** dikirim lewat Brevo HTTP API; `pkg/mailer` diekspos sebagai
  interface `Mailer` agar mudah diganti atau di-mock.
- **Rate limit** dipasang khusus di grup `/api/auth` (5 req/detik, burst 10)
  untuk mencegah brute force.

Menambah fitur baru: buat model + repository + service + dto + handler,
wire repo -> service -> handler di `main.go`,
lalu daftarkan route-nya di `routes/router.go`.

## Setup

```bash
cp .env.example .env
# sesuaikan DB_USER, DB_PASSWORD, DB_NAME, dll.

go mod tidy
```

Buat database PostgreSQL sesuai `DB_NAME` (default `db_go_api`), lalu:

```bash
go run ./cmd/server
# atau: go run cmd/server/main.go
```

Server jalan di `:3000` (atur via `APP_PORT`).

## Menjalankan dengan Docker

Cara tercepat: seluruh stack (API + PostgreSQL + Redis) jalan tanpa perlu
install Go, PostgreSQL, atau Redis di mesin lokal.

```bash
cp .env.example .env
# isi DB_PASSWORD dan JWT_SECRET (minimal 32 karakter) di .env

docker compose up -d --build
```

Akses API di `http://localhost:3000` dan Swagger UI di
`http://localhost:3000/swagger/index.html`.

Perintah lain:

```bash
docker compose logs -f app    # lihat log API
docker compose ps             # status container + health
docker compose down           # stop (data di volume tetap aman)
docker compose down -v        # stop + hapus data PostgreSQL & Redis
```

### Service

| Service    | Image               | Port host (default) | Keterangan                          |
| ---------- | ------------------- | ------------------- | ----------------------------------- |
| `app`      | build dari lokal    | `3000`              | API Go (non-root, healthcheck aktif) |
| `postgres` | `postgres:16-alpine`| `5432`              | volume `postgres_data`              |
| `redis`    | `redis:7-alpine`    | `6379`              | volume `redis_data`, AOF aktif      |

`app` menunggu `postgres` dan `redis` benar-benar sehat
(`depends_on: condition: service_healthy`) sebelum start, jadi tidak ada
error koneksi saat cold start. Migrasi tabel dijalankan otomatis oleh
aplikasi saat startup ←- tidak perlu init script SQL terpisah.

### Host override

`.env` default memakai `DB_HOST=127.0.0.1` dan `REDIS_HOST=localhost` yang
tidak berlaku di dalam container. Compose men-override keduanya ke nama
service (`postgres` / `redis`) via blok `environment`, sehingga `.env` yang
sama tetap bisa dipakai untuk `go run` di lokal maupun di dalam Docker.

### Port bentrok

Jika port default sudah terpakai, ubah via environment variable tanpa
mengedit compose:

```bash
DB_PORT_HOST=5433 REDIS_PORT_HOST=6380 APP_PORT_HOST=3001 docker compose up -d
```

> Di PowerShell, set dulu dengan `$env:DB_PORT_HOST="5433"` sebelum
> menjalankan `docker compose up -d`.

### Catatan mode production

Image sudah default `APP_ENV=production`, tetapi `docker-compose.yml` mengikuti
`APP_ENV` dari `.env` Anda. Untuk benar-benar jalan di mode production, ubah di
`.env`:

```env
APP_ENV=production
DB_SSLMODE=require
```

`config.Validate()` akan fail-fast bila `JWT_SECRET` masih default/kurang dari
32 karakter, atau bila `DB_SSLMODE=disable` dipakai di production.

## API

| Method | Endpoint              | Auth | Keterangan      |
| ------ | --------------------- | ---- | --------------- |
| GET    | `/`                   | -    | Hello World     |
| GET    | `/health`             | -    | Health check DB & Redis |
| GET    | `/api/posts`          | Ya   | List posts      |
| POST   | `/api/posts`          | Ya   | Buat post       |
| GET    | `/api/posts/:id`      | Ya   | Detail post     |
| PUT    | `/api/posts/:id`      | Ya   | Update post     |
| DELETE | `/api/posts/:id`      | Ya   | Hapus post      |
| GET    | `/api/categories`     | Ya   | List categories |
| POST   | `/api/categories`     | Ya   | Buat category   |
| GET    | `/api/categories/:id` | Ya   | Detail category |
| PUT    | `/api/categories/:id` | Ya   | Update category |
| DELETE | `/api/categories/:id` | Ya   | Hapus category  |
| POST   | `/api/auth/*`         | -    | Register, login, refresh, OTP, dsb. |
| GET    | `/swagger/*any`       | -    | Swagger UI      |

Endpoint `/api/*` memerlukan header `Authorization: Bearer <access_token>`.

Body JSON: `{ "title": "...", "content": "..." }` (keduanya required).

## Health check

`GET /health` mem-ping PostgreSQL dan Redis secara nyata (timeout 2 detik),
lalu mengembalikan **200** bila semua sehat atau **503** bila ada dependency
yang down.

```bash
curl http://localhost:3000/health
```

```json
{
  "success": true,
  "message": "Service healthy",
  "data": { "status": "ok", "database": "ok", "redis": "ok" }
}
```

Saat Redis mati, statusnya berubah menjadi `503` dengan
`"status": "degraded"` dan `"redis": "down"` ←- sehingga Docker menandai
container `unhealthy` dan load balancer berhenti mengirim traffic.

## Dokumentasi API (Swagger)

Dokumentasi interaktif tersedia via Swagger UI. Annotations ditulis sebagai
komentar di `cmd/server/main.go` (info API umum) dan
`internal/handlers/post_handler.go` (per endpoint), lalu digenerate menjadi
`docs/` (docs.go, swagger.json, swagger.yaml).

### Cara Menggunakan

Jalankan server seperti biasa, lalu buka Swagger UI di browser:

```
http://localhost:3000/swagger/index.html
```

Dari sana bisa melihat semua endpoint, mencoba request langsung
("Try it out"), dan melihat skema request/response.

### Regenerasi Dokumentasi

Setiap kali annotations Swagger diubah (mis. menambah endpoint baru atau
mengubah DTO), regenerate docs dengan:

```bash
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

> `--parseInternal` diperlukan karena project memakai struktur `internal/`,
> dan `--parseDependency` agar tipe `response.Envelope` yang direferensikan
> di annotations bisa di-resolve.

Alternatif: install `swag` CLI sekali saja, lalu cukup jalankan `swag init ...`:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

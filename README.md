# Zolix Shoe Care Management System

MVP aplikasi operasional cuci sepatu sesuai `PRD.md`: dashboard, order management, customer, service, invoice/tracking drawer, dan API backend Go.

## Run Mode Demo

Mode default memakai in-memory store, cocok untuk development cepat tanpa database.

```bash
go run ./cmd/server
```

Buka:

```text
http://localhost:8080
```

Login default:

```text
email: admin@zolix.test
password: admin123
```

## Run Dengan PostgreSQL

1. Siapkan PostgreSQL lokal.
2. Copy `.env.example` ke `.env`.
3. Sesuaikan `DATABASE_URL`.
4. Jalankan server.

```bash
go run ./cmd/server
```

Contoh `.env`:

```env
PORT=8080
STORE=postgres
DATABASE_URL=postgres://zolix:zolix@localhost:5432/zolix?sslmode=disable
```

Saat `STORE=postgres`, aplikasi akan:

- membuka koneksi PostgreSQL via GORM
- menjalankan `AutoMigrate`
- membuat seed user, customer, service, order, media, dan timeline jika database masih kosong

## Docker Compose PostgreSQL

Jika Docker tersedia:

```bash
docker compose up -d postgres
```

Lalu jalankan server dengan `.env` di atas.

## Docker App Runtime

Build binary ARM64 lebih dulu, lalu jalankan container aplikasi:

```bash
GOOS=linux GOARCH=arm64 go build -o build/zolix-linux-arm64 ./cmd/server
docker compose up -d --build app
```

Default compose memetakan aplikasi ke:

```text
http://localhost:8090
```

## Database Backup

Backup PostgreSQL:

```bash
scripts/backup-postgres.sh
```

Default output:

```text
/data/zolix/backups/zolix-YYYYMMDD-HHMMSS.sql.gz
```

Default retention: 7 hari.

## API

- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`
- `GET /api/v1/dashboard`
- `GET /api/v1/orders`
- `POST /api/v1/orders`
- `GET /api/v1/orders/{id}`
- `PUT /api/v1/orders/{id}`
- `DELETE /api/v1/orders/{id}`
- `GET /api/v1/customers`
- `GET /api/v1/services`
- `POST /api/v1/upload`

Semua endpoint `/api/v1/*` selain login memerlukan header:

```text
Authorization: Bearer <token>
```

Upload foto menggunakan `multipart/form-data`:

```text
order_id=<id order>
type=before|after
file=<jpg|png|webp>
```

Di Docker server, file upload disimpan di:

```text
/data/zolix/uploads
```

## Struktur

```text
cmd/server          entrypoint aplikasi
internal/app        model, HTTP server, in-memory store, PostgreSQL store
web/static          UI dashboard
assets              referensi visual dan gambar demo
```
# zolix-golang

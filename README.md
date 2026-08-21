# Scaean Gate

A centralized Identity and Authorization Provider with OAuth 2.0 Authorization Code + PKCE, group-based application access, independent relying-application sessions, and Kafka-backed session revocation.

App A bernama **Apex**, sedangkan App B bernama **Bolt**.

## Identitas

| Data | Nilai |
| --- | --- |
| Nama | Daniel Arrigo Manurung |
| NIM | 18224031 |

## Panduan Menjalankan Sistem

### Prasyarat

- Docker Engine dengan Docker Compose
- Memori kosong minimal 4 GB untuk PostgreSQL, Kafka, layanan Go, dan tiga frontend

### Konfigurasi

Salin konfigurasi development yang telah disediakan:

```bash
cp .env.example .env
```

Nilai pada `.env.example` dapat langsung digunakan untuk menjalankan sistem secara lokal. Ganti seluruh secret dan gunakan `COOKIE_SECURE=true` sebelum deployment produksi. `.env` diabaikan oleh Git dan tidak boleh di-commit.

### Menjalankan Sistem

Jalankan dari root repositori:

```bash
docker compose up --build
```

Untuk menjalankan di background dan menunggu seluruh health check:

```bash
docker compose up -d --build --wait
docker compose ps
```

### Pembuatan Database, Migrasi, dan Seeder

Migration dan seeder berjalan otomatis saat `docker compose up`; tidak ada perintah tambahan yang diperlukan.

Seeder membuat data berikut:

- `admin@scaean-gate.com` di dalam grup **Admin**
- `testuser@scaean-gate.com` di dalam grup **User**
- OAuth client Apex dan Bolt, redirect URI, serta kebijakan `allow` untuk grup User

Kedua pengguna pada awalnya menggunakan nilai `SEED_USER_PASSWORD` dari `.env`.

Untuk membuat ulang seluruh database dan menjalankan inisialisasi dari kondisi bersih:

```bash
docker compose down -v
docker compose up -d --build --wait
```

> Penghapusan volume akan menghapus seluruh data lokal secara permanen.

### URL Akses

| Komponen | URL | Kegunaan |
| --- | --- | --- |
| Admin Control Panel | <http://localhost:4200> | Profil pusat dan pengelolaan administratif |
| Apex (App A) | <http://localhost:4201> | Relying application independen pertama |
| Bolt (App B) | <http://localhost:4202> | Relying application independen kedua |
| Auth Provider API | <http://localhost:8080> | API autentikasi, OAuth, kebijakan, dan administrasi |
| App A API | <http://localhost:8081> | Callback OAuth Apex dan API sesi lokal |
| App B API | <http://localhost:8082> | Callback OAuth Bolt dan API sesi lokal |
| PostgreSQL | `localhost:5432` | Akses database dari host |
| Kafka | `localhost:9092` | Akses message broker dari host |
| Probe Sync Worker | Port internal `8083` | API health khusus jaringan container |

### Menghentikan Sistem

```bash
docker compose down
```

Gunakan `docker compose down -v` hanya jika volume database juga ingin dihapus.

## Arsitektur dan Alur Request

```mermaid
flowchart LR
    Browser[Browser]
    Control[Angular Admin UI]
    ApexUI[Angular Apex UI]
    BoltUI[Angular Bolt UI]
    IdP[Auth Provider / Gin]
    Apex[App A / Gin]
    Bolt[App B / Gin]
    PG[(PostgreSQL\nsso_db + app_a_db + app_b_db)]
    Kafka[(Apache Kafka)]
    Worker[Sync Worker]

    Browser --> Control --> IdP
    Browser --> ApexUI --> Apex
    Browser --> BoltUI --> Bolt
    Apex <-->|OAuth 2.0 + PKCE| IdP
    Bolt <-->|OAuth 2.0 + PKCE| IdP
    IdP --> PG
    Apex --> PG
    Bolt --> PG
    IdP -->|transactional outbox| Kafka
    Kafka --> Worker
    Worker -->|Bearer shared secret\n/internal/logout| Apex
    Worker -->|Bearer shared secret\n/internal/logout| Bolt
```

### Alur Sign-in dan SSO

1. Apex atau Bolt membuat PKCE verifier, challenge `S256`, dan state OAuth.
2. Browser diarahkan ke `GET /authorize`. Pengguna login jika belum memiliki sesi SSO.
3. Auth Provider memvalidasi pengguna, client, redirect URI, dan kebijakan akses.
4. Auth Provider menerbitkan authorization code sekali pakai ke callback aplikasi.
5. Aplikasi menukarkan code dan verifier melalui `POST /token`.
6. Auth Provider memvalidasi PKCE lalu menerbitkan opaque token.
7. Aplikasi mengambil `/userinfo` dan membuat sesi lokal.
8. Sesi SSO yang sama dapat digunakan di aplikasi lain tanpa login ulang.

### Local Logout

`POST /logout` hanya mencabut sesi lokal. Sesi SSO dan aplikasi lain tetap aktif.

### SSO Logout dan Pencabutan Asinkron

1. Auth Provider mencabut sesi pusat dan token terkait.
2. Event disimpan ke transactional outbox.
3. Outbox mengirim event ke Kafka.
4. Sync Worker memproses event dan delivery setiap aplikasi.
5. Worker memanggil `POST /internal/logout` pada aplikasi terdampak.
6. Aplikasi mencabut sesi lokal secara idempoten.
7. Delivery gagal akan di-retry atau dikirim ke dead-letter topic.

Perubahan password dan akses memakai jalur pencabutan yang sama.

## Keputusan Teknis

- **Opaque token, bukan JWT:** pencabutan dapat berlaku langsung dan token tidak membawa claim lama. Konsekuensinya, validasi harus melalui Auth Provider.
- **Kafka, bukan RabbitMQ:** Kafka menyediakan event log yang tahan lama, replay, consumer group, dan urutan per partisi. Fitur ini cocok untuk pencabutan sesi yang perlu dilacak dan diproses ulang.
- **Shared secret untuk service-to-service:** Sync Worker mengirim `INTERNAL_API_SECRET` sebagai Bearer token saat memanggil `/internal/logout`. Cookie tidak digunakan karena pemanggilnya adalah service, bukan browser pengguna. Payload berisi user dan central session yang harus dicabut, sehingga beberapa sesi tetap dapat diproses saat browser offline. Untuk produksi multi-host, mekanisme ini sebaiknya ditingkatkan ke mTLS atau workload identity.
- **Soft-delete, bukan hard-delete:** data administratif tetap tersedia untuk audit dan relasi historis. Sesi serta token menggunakan status `revoked` atau `expired`.
- **Sesi lokal terpisah:** local logout tidak menghapus sesi SSO atau sesi aplikasi lain. Pencabutan pusat dikirim secara asinkron.
- **Authorization Code + PKCE:** authorization code sekali pakai dan challenge `S256` mencegah code yang dicuri digunakan tanpa verifier.

### Tech Stack dan Versi

| Bagian | Teknologi | Versi |
| --- | --- | --- |
| Backend | Go, Gin, GORM | 1.25.0, 1.12.0, 1.31.2 |
| Frontend | TypeScript, Angular | 5.9.3, 21.2.21 |
| Database | PostgreSQL | 16 |
| Message broker | Apache Kafka, ZooKeeper | 7.5.0 |
| Web server | NGINX | 1.27 |
| Container | Docker, Docker Compose | Compose Specification |

## Daftar Endpoint API

### Auth Provider — `http://localhost:8080`

| Method | Path | Autentikasi | Kegunaan |
| --- | --- | --- | --- |
| GET | `/health` | Publik | Pemeriksaan readiness untuk kompatibilitas |
| GET | `/health/live` | Publik | Liveness proses |
| GET | `/health/ready` | Publik | Readiness PostgreSQL dan Kafka |
| POST | `/login` | Publik | Mengautentikasi dan membuat sesi SSO pusat |
| POST | `/logout` | Sesi pusat | Mencabut sesi SSO pusat |
| POST | `/change-password` | Sesi pusat | Mengubah password dan mencabut akses terkait |
| GET | `/profile` | Sesi pusat | Mengambil profil pusat saat ini |
| GET | `/authorize` | Sesi pusat/parameter OAuth | Memulai atau melanjutkan authorization code flow |
| POST | `/token` | OAuth client credential | Menukarkan code + PKCE verifier dengan opaque token |
| GET | `/userinfo` | Opaque bearer token | Mengambil identitas/profil pemilik token |
| GET | `/admin/users` | Sesi pusat admin | Menampilkan daftar pengguna |
| POST | `/admin/users` | Sesi pusat admin | Membuat pengguna |
| GET | `/admin/users/:id` | Sesi pusat admin | Mengambil detail pengguna |
| PUT | `/admin/users/:id` | Sesi pusat admin | Memperbarui pengguna |
| DELETE | `/admin/users/:id` | Sesi pusat admin | Menghapus pengguna secara soft-delete |
| PATCH | `/admin/users/:id/status` | Sesi pusat admin | Mengaktifkan/menonaktifkan pengguna |
| GET | `/admin/groups` | Sesi pusat admin | Menampilkan daftar grup |
| POST | `/admin/groups` | Sesi pusat admin | Membuat grup |
| GET | `/admin/groups/:id` | Sesi pusat admin | Mengambil grup beserta anggotanya |
| PUT | `/admin/groups/:id` | Sesi pusat admin | Memperbarui grup |
| DELETE | `/admin/groups/:id` | Sesi pusat admin | Menghapus grup secara soft-delete |
| POST | `/admin/groups/:id/users` | Sesi pusat admin | Menambahkan pengguna ke grup |
| DELETE | `/admin/groups/:id/users/:user_id` | Sesi pusat admin | Menghapus pengguna dari grup |
| GET | `/admin/apps` | Sesi pusat admin | Menampilkan daftar aplikasi |
| POST | `/admin/apps` | Sesi pusat admin | Mendaftarkan aplikasi dan menerbitkan secret sekali |
| GET | `/admin/apps/:id` | Sesi pusat admin | Mengambil detail aplikasi |
| PUT | `/admin/apps/:id` | Sesi pusat admin | Memperbarui aplikasi |
| DELETE | `/admin/apps/:id` | Sesi pusat admin | Menghapus aplikasi secara soft-delete |
| POST | `/admin/apps/:id/redirect-uris` | Sesi pusat admin | Menambahkan redirect URI persis |
| DELETE | `/admin/apps/:id/redirect-uris/:uri_id` | Sesi pusat admin | Menghapus redirect URI secara soft-delete |
| GET | `/admin/policies` | Sesi pusat admin | Menampilkan daftar kebijakan akses |
| POST | `/admin/policies` | Sesi pusat admin | Membuat kebijakan akses |
| GET | `/admin/policies/:id` | Sesi pusat admin | Mengambil detail kebijakan akses |
| PUT | `/admin/policies/:id` | Sesi pusat admin | Memperbarui kebijakan akses |
| DELETE | `/admin/policies/:id` | Sesi pusat admin | Menghapus kebijakan akses secara soft-delete |
| GET | `/admin/audit-logs` | Sesi pusat admin | Menampilkan catatan audit keamanan |
| GET | `/admin/events` | Sesi pusat admin | Menampilkan event pencabutan beserta delivery |

### API Apex dan Bolt — port `8081` dan `8082`

Kedua relying application mengimplementasikan path yang sama.

| Method | Path | Autentikasi | Kegunaan |
| --- | --- | --- | --- |
| GET | `/health` | Publik | Pemeriksaan readiness untuk kompatibilitas |
| GET | `/health/live` | Publik | Liveness proses |
| GET | `/health/ready` | Publik | Readiness PostgreSQL lokal |
| GET | `/auth/login` | Publik | Membuat state OAuth/PKCE dan redirect ke Auth Provider |
| GET | `/auth/callback` | OAuth state + authorization code | Menukarkan code dan membuat sesi lokal |
| GET | `/session-status` | Sesi lokal opsional | Mengambil state sesi untuk browser |
| POST | `/internal/logout` | Internal bearer secret | Memproses pencabutan pusat secara idempoten |
| GET | `/me` | Sesi lokal | Mengambil profil lokal yang di-cache |
| GET | `/events` | Sesi lokal | Mengambil state event lokal |
| GET | `/activity` | Sesi lokal | Mengambil aktivitas autentikasi lokal |
| POST | `/logout` | Sesi lokal | Mencabut sesi aplikasi lokal saja |

### Sync Worker — port internal `8083`

| Method | Path | Autentikasi | Kegunaan |
| --- | --- | --- | --- |
| GET | `/health/live` | Internal/publik di dalam jaringan | Liveness proses worker |
| GET | `/health/ready` | Internal/publik di dalam jaringan | Readiness PostgreSQL dan Kafka |

## Fitur Bonus

| Bonus | Status | Implementasi |
| --- | --- | --- |
| B01 | Tidak diimplementasikan | — |
| B02 | Tidak diimplementasikan | — |
| B03 | Selesai | Probe liveness/readiness terpisah dengan pemeriksaan dependency PostgreSQL/Kafka dan health check Compose |
| B04 | Selesai | Penanganan SIGINT/SIGTERM, graceful HTTP shutdown, batas waktu drain, pembatalan outbox, dan penutupan resource di seluruh layanan backend |

## Screenshots

### Apex Sign-in

![Apex sign-in](docs/screenshots/01-apex-sign-in.png)

### Central Identity Login

![Central identity login](docs/screenshots/02-central-login.png)

### Apex Session After OAuth 2.0 + PKCE

![Apex dashboard](docs/screenshots/03-apex-dashboard.png)

### Bolt Using the Existing SSO Session

![Bolt SSO dashboard](docs/screenshots/04-bolt-sso-dashboard.png)

### Administrative Control Panel

![Administrative control panel](docs/screenshots/05-admin-control-panel.png)

### Asynchronously Revoked Application Session

![Revoked application session](docs/screenshots/06-revoked-session.png)

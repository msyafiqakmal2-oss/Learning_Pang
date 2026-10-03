# Peta jalan upgrade EduNexus

| Modul | Berkas | Upgrade yang disarankan |
|---|---|---|
| Penyimpanan | `gateway/store.go`, `jury.go` | Ganti map in-memory dengan PostgreSQL/SQLite; pertahankan nama method. |
| Gamifikasi | `store.go` (bagian "gamifikasi") | Tambah lencana, level kustom, hadiah, event musiman. |
| Kuis | `store.go` | Tipe soal baru (isian, gambar), bank soal, batas waktu, kuis real-time (WebSocket). |
| Juri | `jury.go` | Kriteria dan bobot per lomba, ekspor hasil PDF/CSV, anonimisasi tim. |
| Auth | `auth.go` | Refresh token, cookie HttpOnly, rate limit login, verifikasi guru. |
| Layanan Rust | `hasher/src/main.rs` | Simpan audit ke disk/DB, metrik Prometheus. |
| Frontend | `web/` | Pindah ke React/Svelte; kontrak API `/api/*` tetap. |

## Ide lintas bahasa
- **Python** (`analytics/`): laporan belajar dan rekomendasi materi dari log audit.
- **TypeScript**: frontend bertipe dengan Vite.
- **Node/Deno**: worker notifikasi email/WhatsApp.

## Endpoint
| Metode | Path | Akses |
|---|---|---|
| POST | `/api/auth/register`, `/api/auth/login` | publik |
| GET | `/api/me`, `/api/stats`, `/api/leaderboard` | login |
| GET | `/api/classes` | login |
| POST/DELETE | `/api/classes`, `/api/classes/{id}` | guru |
| POST | `/api/classes/join` | siswa |
| GET/POST | `/api/classes/{id}/quizzes` | login / guru |
| DELETE | `/api/quizzes/{id}` | guru |
| POST | `/api/quizzes/{id}/attempt` | siswa |
| GET/POST | `/api/classes/{id}/posts` | login (juri hanya baca) |
| POST/DELETE | `/api/posts/{id}/like`, `/api/posts/{id}` | login / penulis atau guru kelas |
| GET/POST | `/api/projects` | login / siswa |
| DELETE | `/api/projects/{id}` | pemilik atau guru |
| POST | `/api/projects/{id}/score` | juri |
| GET | `/api/audit` | guru, juri |

# EduNexus — Ekosistem Kolaborasi & Gamifikasi Pendidikan

Belajar jadi petualangan: kuis, misi harian, , level, streak, lencana, kelas kolaboratif,
dan modul **penilaian juri** untuk memilih karya terbaik.

```
nexusstack/
├── gateway/   Go    – API, JWT, 3 peran (siswa/guru/juri), gamifikasi, penilaian juri
├── hasher/    Rust  – hashing Argon2id + audit log transparan
├── web/       JS    – antarmuka (login, beranda, kelas, kuis, peringkat, karya, panduan)
├── docs/      panduan juri & upgrade
└── docker-compose.yml, Makefile
```

## Menjalankan
```bash
make dev                      # Go saja → http://localhost:8080
# atau lengkap (Go + Rust):   make rust   |   make go
# atau Docker:                cp .env.example .env && docker compose up --build
```
Data demo terisi otomatis (matikan dengan `SEED_DEMO=false`).

## Akun demo (kata sandi `demo12345`)
| Peran | Nama pengguna |
|---|---|
| Guru | `bu_sari` |
| Siswa | `rina`, `putri`, `dimas`, `bayu`, `citra` |
| Juri | `juri1`, `juri2` |

Kode kelas demo: `MTK7A`, `ENG8B`. Kode mendaftar sebagai juri: `JURI2026` (ubah lewat `JURY_CODE`).

## Fitur utama
- **Gamifikasi:** XP, gelar level (Pemula → Legenda), streak, 3 misi harian, 8 lencana, bonus kecepatan ⚡, papan peringkat.
- **Kolaborasi:** gabung kelas dengan kode, diskusi + apresiasi (like), Tantangan Kelas (total XP kelas).
- **Kuis:** guru membuat kuis sendiri; siswa dapat timer + pembahasan; XP hanya pada percobaan pertama; guru melihat statistik (rata-rata, soal tersulit, hasil siswa).
- **Karya & Juri:** siswa mengirim karya; juri menilai 5 kriteria berbobot; podium juara otomatis, sertifikat juara yang bisa dicetak, rekap nilai CSV.
- **Keamanan:** JWT, RBAC per peran, Argon2id (Rust), validasi server, UI anti-XSS, log audit.

## Penilaian juri
Skor akhir 0-100 = rata-rata nilai semua juri, dengan bobot:
Inovasi 25% · Dampak Pendidikan 25% · Kolaborasi 20% · Desain & UX 15% · Kualitas Teknis 15%.

## Catatan
- Data disimpan permanen di `edunexus.json` (Docker: volume `edudata`). Hapus berkas/volume untuk mengulang dari awal. Set `DATA_FILE=off` untuk mematikan.
- Ganti `JWT_SECRET` dan `JURY_CODE` sebelum dipakai sungguhan.

# Mengumpulkan ke lomba & memudahkan juri

## 1. Taruh online (pilih satu)
Juri tidak akan menjalankan Docker. Beri mereka **satu tautan https**.

| Cara | Cocok untuk | Catatan |
|---|---|---|
| **A. VPS kecil + Docker Compose** | Paling stabil | `git clone`, isi `.env`, `docker compose up -d --build`. Pasang domain/HTTPS (mis. Caddy) bila perlu. |
| **B. Layanan hosting container** (Render, Railway, Fly.io, dsb.) | Cepat, tanpa server sendiri | Deploy dari GitHub memakai `Dockerfile.single` (Go + Rust dalam satu container). Cek dulu syarat paket gratis: sebagian "tidur" saat sepi sehingga bukaan pertama lambat. |
| **C. Tunnel dari laptop** (Cloudflare Tunnel / ngrok) | Demo langsung saat sesi juri | Laptop harus menyala dan terhubung internet selama juri menilai. |

Uji sebelum dikirim: `docker build -f Dockerfile.single -t edunexus . && docker run -p 8080:8080 edunexus`.

### Variabel lingkungan wajib di server publik
| Variabel | Nilai |
|---|---|
| `JWT_SECRET` | string acak panjang (minimal 32 karakter) |
| `JURY_CODE` | kode rahasia, bagikan hanya ke panitia/juri |
| `SEED_DEMO` | `true` agar akun demo dan data contoh selalu ada |
| `DATA_FILE` | `/data/edunexus.json` (pasang volume ke `/data` bila tersedia) |

> Akun demo (kata sandi `demo12345`) bersifat publik. Bagikan tautan hanya ke panitia/juri, dan jangan taruh data asli di server demo.

## 2. Tautan untuk juri (tanpa mengetik)
Setelah online, ganti `DOMAIN` dengan alamat Anda:
- Juri: `https://DOMAIN/?demo=juri`
- Siswa: `https://DOMAIN/?demo=siswa`
- Guru: `https://DOMAIN/?demo=guru`

Juri langsung masuk, lalu melihat **Tur juri 5 menit** di Beranda.

## 3. Checklist sebelum mengumpulkan
- [ ] Tautan demo terbuka di HP dan laptop, di jaringan berbeda (bukan hanya wifi sendiri).
- [ ] Dibuka sekali beberapa menit sebelum penilaian agar server tidak "tidur".
- [ ] Menu **Panduan → Cek sistem** menunjukkan Go dan Rust sehat.
- [ ] Repositori GitHub **publik**, README di atas berisi tautan demo + akun demo + screenshot.
- [ ] Video demo 2-3 menit (alur: Siswa → kuis → Juri menilai → podium → sertifikat).
- [ ] `JWT_SECRET` dan `JURY_CODE` sudah diganti, file `.env` tidak ikut ke GitHub.
- [ ] Kirim sesuai format panitia: biasanya tautan demo + tautan repo + video + dokumen singkat.

## 4. Lembar singkat untuk juri (salin ke dokumen/PDF)
**Cara membuka:** klik tautan juri → langsung masuk. Tidak perlu daftar atau memasang apa pun.

**Cara menilai (5-10 menit):** ikuti *Tur juri* di Beranda, lalu beri skor 1-10 per kriteria:

| Kriteria (bobot) | Lihat di mana | Pertanyaan pemandu |
|---|---|---|
| Inovasi (25%) | Beranda siswa, Peringkat | Apakah gamifikasi (XP, level, misi, lencana, bonus kecepatan) terasa baru dan memotivasi? |
| Dampak Pendidikan (25%) | Kelas → Kuis (siswa), Statistik (guru) | Apakah kuis dan statistik benar-benar membantu belajar dan mengajar? |
| Kolaborasi (20%) | Kelas → Diskusi, Tantangan Kelas | Apakah ada insentif bekerja sama? |
| Desain & UX (15%) | Seluruh halaman, buka di HP | Mudah dipahami, menarik, rapi di layar kecil? |
| Kualitas Teknis (15%) | Panduan → Cek sistem, log audit | Cepat, stabil, aman (login, peran, audit)? |

**Proses pembukaan:** nilai juga kecepatan buka halaman, kemudahan masuk, dan hasil *Cek sistem* (Panduan Juri).

#!/usr/bin/env python3
"""Membuat kelas + kuis (5 soal) untuk guru di EduNexus lewat API.
Jalankan: python3 seed_teddy.py
Server harus sedang berjalan. Kelas yang namanya sudah ada akan dilewati."""
import getpass, json, os, sys, urllib.request, urllib.error

BASE = os.environ.get("EDU_URL", "http://localhost:8090").rstrip("/")

# Setiap soal: (pertanyaan, [pilihan...], indeks jawaban benar mulai dari 0)
KELAS = {
    "Pemodelan Optimasi": ("Dasar Pemrograman Linear", [
        ("Dalam pemrograman linear, fungsi yang ingin dimaksimalkan atau diminimalkan disebut...",
         ["Fungsi kendala", "Fungsi objektif", "Variabel keputusan", "Solusi dasar"], 1),
        ("Metode grafik pada pemrograman linear umumnya dipakai untuk berapa variabel keputusan?",
         ["Satu", "Dua", "Lima", "Sepuluh"], 1),
        ("Daerah yang memenuhi semua kendala disebut...",
         ["Daerah layak", "Daerah optimum mutlak", "Daerah tak terbatas", "Daerah simpleks"], 0),
        ("Metode iteratif untuk menyelesaikan pemrograman linear dengan banyak variabel adalah...",
         ["Newton-Raphson", "Simpleks", "Trapesium", "Gauss-Seidel"], 1),
        ("Nilai maksimum Z = 3x + 2y pada titik sudut (0,0), (4,0), (2,3), dan (0,4) adalah...",
         ["8", "12", "14", "18"], 1),
    ]),
    "Komputasi Awan": ("Konsep Dasar Cloud Computing", [
        ("Model layanan cloud yang menyediakan server dan penyimpanan virtual disebut...",
         ["SaaS", "PaaS", "IaaS", "XaaS"], 2),
        ("Gmail dan Google Docs termasuk model layanan...",
         ["IaaS", "PaaS", "SaaS", "On-premise"], 2),
        ("Kemampuan sumber daya bertambah atau berkurang otomatis sesuai beban disebut...",
         ["Elastisitas", "Redundansi", "Enkripsi", "Latensi"], 0),
        ("Teknologi yang memungkinkan satu server fisik menjalankan banyak mesin virtual adalah...",
         ["Kompresi", "Virtualisasi", "Fragmentasi", "Replikasi"], 1),
        ("Model deployment cloud yang dipakai oleh satu organisasi saja disebut cloud...",
         ["Publik", "Privat", "Hibrida", "Komunitas"], 1),
    ]),
    "Statistika": ("Statistika Deskriptif", [
        ("Rata-rata dari data 4, 6, 8, 10, 12 adalah...", ["6", "8", "10", "9"], 1),
        ("Median dari data 2, 4, 6, 8 adalah...", ["4", "5", "6", "7"], 1),
        ("Nilai yang paling sering muncul dalam data disebut...",
         ["Mean", "Median", "Modus", "Rentang"], 2),
        ("Ukuran sebaran data terhadap rata-rata adalah...",
         ["Modus", "Simpangan baku", "Median", "Kuartil"], 1),
        ("Distribusi normal baku memiliki rata-rata dan simpangan baku sebesar...",
         ["0 dan 1", "1 dan 0", "0 dan 0", "1 dan 1"], 0),
    ]),
    "Metode Numerik": ("Akar Persamaan dan Integrasi Numerik", [
        ("Metode mencari akar dengan membagi dua interval berulang disebut...",
         ["Bagi dua (bisection)", "Simpson", "Euler", "Gauss-Jordan"], 0),
        ("Metode Newton-Raphson memakai informasi dari...",
         ["Integral fungsi", "Turunan fungsi", "Invers matriks", "Interpolasi Lagrange"], 1),
        ("Metode Trapesium dipakai untuk mendekati...",
         ["Akar persamaan", "Integral tentu", "Determinan", "Nilai eigen"], 1),
        ("Galat yang timbul karena membulatkan bilangan disebut galat...",
         ["Pemotongan", "Pembulatan", "Relatif", "Sistematis"], 1),
        ("Metode Euler digunakan untuk menyelesaikan...",
         ["Persamaan diferensial biasa", "Sistem pertidaksamaan", "Statistik deskriptif", "Pencarian string"], 0),
    ]),
}


def call(method, path, token=None, body=None):
    req = urllib.request.Request(
        BASE + path, method=method,
        data=json.dumps(body).encode() if body is not None else None,
        headers={"Content-Type": "application/json", **({"Authorization": "Bearer " + token} if token else {})})
    try:
        with urllib.request.urlopen(req) as r:
            d = r.read()
            return json.loads(d) if d else None
    except urllib.error.HTTPError as e:
        msg = json.loads(e.read() or b"{}").get("error", e.reason)
        sys.exit(f"Gagal ({e.code}): {msg}")
    except urllib.error.URLError:
        sys.exit(f"Tidak bisa terhubung ke {BASE}. Pastikan server berjalan (docker compose ps).")


def main():
    user = os.environ.get("EDU_USER") or input("Nama pengguna guru [Pak Ahmad Teddy]: ").strip() or "Pak Ahmad Teddy"
    pw = os.environ.get("EDU_PASS") or getpass.getpass("Kata sandi: ")
    token = call("POST", "/api/auth/login", body={"username": user, "password": pw})["token"]
    me = call("GET", "/api/me", token)
    if me["role"] != "guru":
        sys.exit(f"Akun '{me['username']}' berperan {me['role']}, harus guru.")
    ada = {c["name"] for c in call("GET", "/api/classes", token)}
    for nama, (judul, soal) in KELAS.items():
        if nama in ada:
            print(f"- {nama}: sudah ada, dilewati")
            continue
        c = call("POST", "/api/classes", token, {"name": nama})
        call("POST", f"/api/classes/{c['id']}/quizzes", token, {
            "title": judul, "xp_reward": 60,
            "questions": [{"text": t, "options": o, "answer": a} for t, o, a in soal]})
        print(f"- {nama}: dibuat (kode kelas {c['code']}) + kuis '{judul}' ({len(soal)} soal)")
    print("Selesai. Muat ulang halaman EduNexus untuk melihat hasilnya.")


main()

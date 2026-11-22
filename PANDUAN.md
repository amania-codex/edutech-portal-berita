# Panduan menjalankan portal berita

Arsip berita lama (edukator.id) sudah ikut di repo ini:

- teks berita: `backend/database/seed/berita.json.gz`, **dimuat otomatis** saat backend pertama kali dinyalakan pada database yang belum berisi arsip;
- gambar: `backend/public/uploads/wp/`, disajikan backend lewat `/uploads/wp/...`.

## Setelah `git pull`

Butuh Go (versi di `backend/go.mod`), Node 22.12+, dan MySQL 8.

```bash
# 1. Backend (port 3000). Tabel dibuat dan berita dimuat otomatis.
cd backend
go run .
# tunggu: "SeedNews: 2817 berita arsip dimuat"

# 2. Frontend, di terminal lain. Backend harus tetap menyala saat build.
cd frontend
npm ci
npx astro build
npx astro preview      # http://localhost:4321
```

Memastikan semua berita dan gambar tampil:

```bash
cd backend && MIN_NEWS=2817 node tools/check-images.mjs
```

Windows dengan Smart App Control aktif memblokir program Go buatan sendiri. Jalankan backend di server Linux (atau WSL).

## Produksi

1. Variabel backend ada di `backend/.env.example`. `JWT_SECRET` wajib diisi, dan `CORS_ORIGINS` berisi alamat frontend.
2. Akun `admin@edutech.id` / `admin123` hanya dibuat di luar mode produksi. Bila pernah terbuat, **segera ganti passwordnya**; repo ini publik.
3. nginx di depan backend: `proxy_set_header X-Real-IP $remote_addr;` agar pembatas login menghitung per pengunjung.
4. Build frontend dengan alamat API: `PUBLIC_API_URL=https://api.contoh.id npx astro build`, lalu unggah `frontend/dist/`.
5. Arahkan halaman tak ditemukan ke `404.html`. Halaman itu mengalihkan alamat WordPress lama ke alamat baru.
6. Frontend statis: **build ulang setiap ada berita baru** agar muncul di situs.

## Perkakas (`backend/tools/`)

| File | Guna |
|---|---|
| `import-wp.php` | Impor dari dump WordPress ke MySQL |
| `export-seed.php` | Membuat ulang `berita.json.gz` (hanya berita terbit) |
| `check-images.mjs` | Memeriksa semua berita dan gambar dari backend yang menyala |
| `smoke-test.mjs` | Menguji seluruh fitur API; hanya untuk localhost, bukan database produksi |

Mematikan pemuatan arsip otomatis: `SEED_NEWS=false`.

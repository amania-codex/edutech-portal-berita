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

1. Salin `backend/.env.example` menjadi `backend/.env` lalu isi. `PORT` harus port yang belum dipakai aplikasi lain.
2. Pasang semuanya:

   ```bash
   bash deploy/deploy.sh cek      # hanya memeriksa, tidak mengubah apa pun
   bash deploy/deploy.sh          # backup database, backend (pm2), periksa berita, build situs
   ```

   Database dicadangkan ke `~/backup-edutech/` sebelum apa pun diubah. Bila backend baru gagal menyala, versi sebelumnya dinyalakan lagi.
3. nginx: arahkan root situs ke `frontend/dist`, lalu tempel isi `deploy/nginx-lokasi.conf` (sesuaikan port).
4. Agar backend hidup lagi setelah server restart, tambahkan ke `crontab -e`:
   `@reboot PATH=<isi dari: echo $PATH> pm2 resurrect`
5. **Setiap ada berita baru** jalankan `bash deploy/deploy.sh frontend`, karena situsnya statis.

Jangan jalankan `backend/server_backend`: itu binary lama yang punya celah keamanan.

## Perkakas (`backend/tools/`)

| File | Guna |
|---|---|
| `import-wp.php` | Impor dari dump WordPress ke MySQL |
| `export-seed.php` | Membuat ulang `berita.json.gz` (hanya berita terbit) |
| `check-images.mjs` | Memeriksa semua berita dan gambar dari backend yang menyala |
| `smoke-test.mjs` | Menguji seluruh fitur API; hanya untuk localhost, bukan database produksi |

Mematikan pemuatan arsip otomatis: `SEED_NEWS=false`.

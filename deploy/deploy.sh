#!/usr/bin/env bash
# Memasang atau memperbarui portal di server Linux.
#
#   bash deploy/deploy.sh            semuanya (backend lalu frontend)
#   bash deploy/deploy.sh backend    hanya backend
#   bash deploy/deploy.sh frontend   hanya halaman situs; jalankan setiap ada berita baru
#   bash deploy/deploy.sh cek        hanya memeriksa kesiapan, tidak mengubah apa pun
#
# Pengaturan dibaca dari backend/.env (contoh: backend/.env.example).
# Tidak menyentuh aplikasi lain, nginx, maupun database selain milik portal ini.
set -Eeuo pipefail
cd "$(dirname "$0")/.."
ROOT=$PWD
APP=edutech-backend
ENV_FILE=$ROOT/backend/.env
MODE=${1:-semua}

say()  { printf '\n\033[1;36m== %s\033[0m\n' "$*"; }
ok()   { printf '   ✓ %s\n' "$*"; }
die()  { printf '\n\033[1;31m✗ %s\033[0m\n' "$*" >&2; exit 1; }
trap 'printf "\n\033[1;31m✗ Berhenti di baris %s. Situs yang sedang tayang tidak diganti.\033[0m\n" "$LINENO" >&2' ERR

# nilai dari .env: baris terakhir menang, tanpa kutip dan tanpa CR
env_get() {
  local v
  v=$(sed -nE "s/^[[:space:]]*(export[[:space:]]+)?$1[[:space:]]*=[[:space:]]*(.*)$/\2/p" "$ENV_FILE" | tail -n1 | tr -d '\r' | sed -E 's/[[:space:]]+$//')
  case $v in \"*\") v=${v#\"}; v=${v%\"} ;; \'*\') v=${v#\'}; v=${v%\'} ;; esac
  printf '%s' "$v"
}

port_busy() { ss -ltnH 2>/dev/null | awk '{print $4}' | grep -qE "[:.]$1\$"; }
running()   { local pid; pid=$(pm2 pid "$APP" 2>/dev/null | grep -E '^[0-9]+$' | tail -n1 || true); [ -n "$pid" ] && [ "$pid" != 0 ]; }
# Jawaban harus daftar kategori (JSON array), supaya tidak tertukar dengan aplikasi lain di port yang sama
healthy()   { [ "$(curl -s -m 5 "http://127.0.0.1:$PORT/api/categories" | head -c1)" = '[' ]; }

check() {
  say "Memeriksa kesiapan"
  local c problems=()
  for c in go node npm pm2 curl mysqldump gzip ss; do command -v "$c" >/dev/null || problems+=("program '$c' belum terpasang"); done
  [ -f "$ENV_FILE" ] || die "File backend/.env belum ada. Salin dari backend/.env.example lalu isi."
  PORT=$(env_get PORT); SITE_URL=$(env_get SITE_URL); SITE_URL=${SITE_URL%/}
  for c in APP_ENV PORT SITE_URL DB_NAME DB_USER DB_PASSWORD JWT_SECRET CORS_ORIGINS; do
    [ -n "$(env_get $c)" ] || problems+=("$c belum diisi di backend/.env")
  done
  [ "$(env_get APP_ENV)" = production ] || problems+=("APP_ENV harus 'production'")
  [ "$(env_get JWT_SECRET | wc -c)" -ge 32 ] || problems+=("JWT_SECRET minimal 32 karakter (buat dengan: openssl rand -hex 32)")
  case $SITE_URL in http://*|https://*) ;; *) problems+=("SITE_URL harus diawali https://") ;; esac
  if command -v pm2 >/dev/null; then pm2 ping >/dev/null 2>&1 || true; fi
  if [[ ! $PORT =~ ^[0-9]+$ ]]; then problems+=("PORT harus angka")
  elif command -v pm2 >/dev/null && ! running && port_busy "$PORT"; then problems+=("port $PORT sudah dipakai aplikasi lain; pilih PORT lain di backend/.env"); fi
  if [ ${#problems[@]} -gt 0 ]; then
    for c in "${problems[@]}"; do printf '   ✗ %s\n' "$c"; done
    die "Perbaiki dulu hal di atas. Belum ada yang diubah."
  fi
  ok "backend di 127.0.0.1:$PORT, situs $SITE_URL"
}

backup_db() {
  say "Backup database"
  local dir=$HOME/backup-edutech file
  mkdir -p "$dir"; chmod 700 "$dir"
  file=$dir/db-$(date +%Y%m%d-%H%M%S).sql.gz
  ( umask 077; MYSQL_PWD=$(env_get DB_PASSWORD) mysqldump --single-transaction --no-tablespaces --default-character-set=utf8mb4 \
      -h "$(env_get DB_HOST | grep . || echo 127.0.0.1)" -P "$(env_get DB_PORT | grep . || echo 3306)" \
      -u "$(env_get DB_USER)" "$(env_get DB_NAME)" | gzip > "$file" ) || { rm -f "$file"; die "Backup gagal. Periksa DB_NAME, DB_USER, dan DB_PASSWORD di backend/.env. Belum ada yang diubah."; }
  ok "$file ($(du -h "$file" | cut -f1))"
}

start_backend() {
  if pm2 describe "$APP" >/dev/null 2>&1; then pm2 restart "$APP" --update-env >/dev/null
  else pm2 start ./bin/edutech-backend --name "$APP" --interpreter none --cwd "$ROOT/backend" >/dev/null; fi
  local i; for i in $(seq 1 60); do healthy && return 0; sleep 1; done
  return 1
}

deploy_backend() {
  backup_db
  say "Membangun backend"
  cd "$ROOT/backend"
  mkdir -p bin
  CGO_ENABLED=0 go build -o bin/edutech-backend.baru .
  ok "selesai"
  say "Menyalakan backend"
  [ -f bin/edutech-backend ] && cp -f bin/edutech-backend bin/edutech-backend.lama
  mv -f bin/edutech-backend.baru bin/edutech-backend
  if ! start_backend; then
    pm2 logs "$APP" --nostream --lines 30 || true
    if [ -f bin/edutech-backend.lama ]; then
      mv -f bin/edutech-backend.lama bin/edutech-backend
      pm2 restart "$APP" >/dev/null && echo "   versi sebelumnya dinyalakan lagi"
    else pm2 delete "$APP" >/dev/null 2>&1 || true; fi
    die "Backend tidak mau menyala. Lihat log di atas."
  fi
  pm2 save >/dev/null
  rm -f bin/edutech-backend.lama
  ok "menyala, $(curl -s -D- -o /dev/null "http://127.0.0.1:$PORT/api/news?limit=1" | tr -d '\r' | awk 'tolower($1)=="x-total-count:"{print $2}') berita terbit"
  cd "$ROOT"
}

check_news() {
  say "Memeriksa semua berita dan gambar"
  BASE_URL=http://127.0.0.1:$PORT node backend/tools/check-images.mjs || echo "   ! sebagian gambar bermasalah (lihat di atas), pemasangan tetap dilanjutkan"
}

deploy_frontend() {
  healthy || die "Backend di port $PORT tidak menjawab. Jalankan dulu: bash deploy/deploy.sh backend"
  say "Membangun halaman situs"
  cd "$ROOT/frontend"
  rm -rf dist.baru
  npm ci --no-audit --no-fund >/dev/null
  # Data diambil langsung dari backend lokal; alamat di halaman tetap SITE_URL
  API_INTERNAL_URL=http://127.0.0.1:$PORT PUBLIC_API_URL=$SITE_URL npx astro build --outDir dist.baru > "$ROOT/frontend/build.log" 2>&1 || { tail -n 30 build.log; die "Build gagal. Situs yang sedang tayang tidak diganti."; }
  [ -f dist.baru/index.html ] && [ -f dist.baru/404.html ] || die "Hasil build tidak lengkap. Situs yang sedang tayang tidak diganti."
  rm -rf dist.lama
  [ -e dist ] && mv dist dist.lama
  mv dist.baru dist
  rm -rf dist.lama build.log
  ok "$(find dist/berita -name index.html | wc -l) halaman berita di $ROOT/frontend/dist"
  cd "$ROOT"
}

check
case $MODE in
  cek) ok "siap" ;;
  backend) deploy_backend; check_news ;;
  frontend) deploy_frontend ;;
  semua) deploy_backend; check_news; deploy_frontend ;;
  *) die "Pilihan tidak dikenal: $MODE (semua | backend | frontend | cek)" ;;
esac
say "Selesai"

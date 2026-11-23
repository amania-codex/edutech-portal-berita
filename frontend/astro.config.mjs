// @ts-check
import { defineConfig } from 'astro/config';
import node from '@astrojs/node';

// Halaman dibuat saat dibuka (server-side rendering): tambah, ubah, dan hapus berita
// langsung terlihat tanpa build ulang. Jalankan: node dist/server/entry.mjs
export default defineConfig({
  output: 'server',
  adapter: node({ mode: 'standalone' }),
  vite: {
    build: {
      cssMinify: 'esbuild',
    },
  },
  server: {
    port: 4321,
    host: true,
  },
});

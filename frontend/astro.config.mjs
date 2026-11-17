// @ts-check
import { defineConfig } from 'astro/config';

// https://astro.build/config
export default defineConfig({
  output: 'static',
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

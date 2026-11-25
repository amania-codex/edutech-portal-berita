// Uji semua form lewat browser sungguhan (Chrome headless, tanpa dependensi).
//
//   BASE_URL=http://localhost:8099 ADMIN_PASSWORD=... node tools/e2e-forms.mjs
//
// Mengisi dan mengirim setiap form seperti pengguna, lalu memeriksa hasilnya langsung tampil
// di situs. Membuat lalu menghapus data uji: hanya untuk server uji, bukan situs produksi.
import { spawn } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

const BASE = process.env.BASE_URL || 'http://localhost:8099';
if (!/^https?:\/\/(localhost|127\.0\.0\.1)(:|\/|$)/.test(BASE)) {
  console.error('Ditolak: hanya untuk server uji di localhost.');
  process.exit(2);
}
const ADMIN = { email: process.env.ADMIN_EMAIL || 'admin@edutech.id', password: process.env.ADMIN_PASSWORD || 'admin123' };
const CHROME = process.env.CHROME_BIN || (process.platform === 'win32' ? 'C:/Program Files/Google/Chrome/Application/chrome.exe' : 'google-chrome');
const run = Date.now().toString(36);
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'e2e-'));
const image = path.join(tmp, 'foto.png');
fs.writeFileSync(image, Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==', 'base64'));

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const chrome = spawn(CHROME, ['--headless=new', '--no-sandbox', '--disable-gpu', '--no-first-run', '--remote-debugging-port=9334', `--user-data-dir=${tmp}/profil`, 'about:blank'], { stdio: 'ignore' });
let targets;
for (let i = 0; i < 80 && !targets?.length; i++) {
  try { targets = (await (await fetch('http://127.0.0.1:9334/json')).json()).filter((t) => t.type === 'page'); } catch { /* belum siap */ }
  await sleep(250);
}
const ws = new WebSocket(targets[0].webSocketDebuggerUrl);
await new Promise((r) => { ws.onopen = r; });
let seq = 0;
const pending = new Map();
const jsErrors = [];
let onLoad = null;
ws.onmessage = ({ data }) => {
  const m = JSON.parse(data);
  if (m.id) { pending.get(m.id)?.(m); pending.delete(m.id); return; }
  if (m.method === 'Runtime.exceptionThrown') jsErrors.push(m.params.exceptionDetails.exception?.description?.split('\n')[0] || m.params.exceptionDetails.text);
  if (m.method === 'Page.loadEventFired') onLoad?.();
  if (m.method === 'Page.javascriptDialogOpening') send('Page.handleJavaScriptDialog', { accept: true });
};
const send = (method, params = {}) => new Promise((r) => { const id = ++seq; pending.set(id, r); ws.send(JSON.stringify({ id, method, params })); });
async function ev(expression) {
  const r = await send('Runtime.evaluate', { expression, awaitPromise: true, returnByValue: true });
  if (r.result?.exceptionDetails) return `GALAT: ${(r.result.exceptionDetails.exception?.description || r.result.exceptionDetails.text).split('\n')[0]}`;
  return r.result?.result?.value;
}
async function go(p) {
  const raw = new URL(BASE + p).pathname;
  const want = raw.length > 1 && raw.endsWith('/') ? raw.slice(0, -1) : raw;
  await send('Page.navigate', { url: BASE + p });
  // tunggu halaman TUJUAN selesai dimuat (bukan load event halaman sebelumnya)
  for (let t = 0; t < 15000; t += 100) {
    const ready = await ev(`(location.pathname.endsWith('/') && location.pathname.length > 1 ? location.pathname.slice(0, -1) : location.pathname) === ${JSON.stringify(want)} && document.readyState === 'complete'`);
    if (ready === true) break;
    await sleep(100);
  }
  await sleep(150);
}
// Keadaan halaman saat sebuah langkah gagal: alamat dan pesan yang sedang terlihat
const STATE = `(() => { const seen = [...document.querySelectorAll('[id*=error],[id*=msg],[id$=-message],.alert,.comment-msg,.toast,#reg-success,#login-success')]
  .filter((e) => e.offsetParent && e.textContent.trim()).map((e) => '#' + e.id + ': ' + e.textContent.trim().slice(0, 120));
  return location.pathname + location.search + ' | ' + (seen.join(' ; ') || document.title); })()`;
let lastState = null;
async function until(expr, ms = 10000) {
  for (let t = 0; t < ms; t += 200) { if ((await ev(expr)) === true) return true; await sleep(200); }
  lastState = await ev(STATE);
  return false;
}
const js = (v) => JSON.stringify(v);
const fill = (sel, value) => ev(`(() => { const e = document.querySelector(${js(sel)}); e.value = ${js(value)}; e.dispatchEvent(new Event('input', { bubbles: true })); e.dispatchEvent(new Event('change', { bubbles: true })); return true; })()`);
const click = (sel) => ev(`(() => { document.querySelector(${js(sel)}).click(); return true; })()`);
const submit = (sel) => ev(`(() => { document.querySelector(${js(sel)}).requestSubmit(); return true; })()`);
const text = (sel) => ev(`document.querySelector(${js(sel)})?.textContent?.trim() ?? ''`);
const path_ = () => ev('location.pathname');
async function upload(sel) {
  const { result: { root } } = await send('DOM.getDocument', { depth: 0 });
  const { result: { nodeId } } = await send('DOM.querySelector', { nodeId: root.nodeId, selector: sel });
  await send('DOM.setFileInputFiles', { nodeId, files: [image] });
}
const api = (p, init) => fetch(BASE + '/api' + p, init);
const pageHas = async (p, needle) => { const r = await fetch(BASE + p); return { status: r.status, has: (await r.text()).includes(needle) }; };

async function eventually(path, needle, present = true, ms = 15000) {
  for (let t = 0; t < ms; t += 1000) {
    const r = await pageHas(path, needle);
    if (r.has === present) return true;
    await sleep(1000);
  }
  return false;
}

let pass = 0;
const failures = [];
function check(name, ok, info) {
  if (!ok && info === undefined && lastState) info = lastState;
  lastState = null;
  if (ok) pass++; else failures.push(`${name}${info === undefined ? '' : ' → ' + JSON.stringify(info).slice(0, 300)}`);
  console.log(`${ok ? '✓' : '✗'} ${name}${ok ? '' : '  ← ' + JSON.stringify(info ?? '').slice(0, 300)}`);
}

await send('Page.enable');
await send('Runtime.enable');
await send('DOM.enable');
await send('Emulation.setDeviceMetricsOverride', { width: 1300, height: 950, deviceScaleFactor: 1, mobile: false });

async function login(email, password) {
  await go('/login');
  await fill('#email', email);
  await fill('#password', password);
  await submit('#login-form');
  return until(`!location.pathname.startsWith('/login')`, 10000);
}
async function logout() {
  if (!(await until(`!!document.getElementById('logout-btn')`, 5000))) return false;
  await click('#logout-btn');
  return until(`location.pathname === '/'`);
}

let eventRow = '';
try {
  // ================= ADMIN =================
  check('login: password salah ditolak', await (async () => { await go('/login'); await fill('#email', ADMIN.email); await fill('#password', 'salah-sekali'); await submit('#login-form'); return until(`getComputedStyle(document.getElementById('login-error')).display !== 'none'`); })());
  check('login admin lewat form', await login(ADMIN.email, ADMIN.password) && (await path_()).startsWith('/dashboard'));
  check('dashboard: angka statistik termuat', await until(`Number(document.getElementById('total-news').textContent.replace(/\\D/g, '')) > 0`));

  // --- tulis berita
  const title = `Berita Uji Form ${run}`;
  const slug = `berita-uji-form-${run}`;
  await go('/dashboard/news/create/');
  await until(`document.getElementById('category').options.length > 2`);
  await ev(`(() => { const t = document.getElementById('post-title'); t.textContent = ${js(title)}; t.dispatchEvent(new Event('input')); document.querySelector('.block-para').innerHTML = 'Paragraf pertama berita uji.'; return true; })()`);
  check('tulis: slug otomatis dari judul', (await ev(`document.getElementById('slug').value`)) === slug);
  await upload('#feat-file-input');
  check('tulis: unggah gambar utama', await until(`document.getElementById('image-url').value.includes('/uploads/')`));
  await fill('#category', 'Pendidikan');
  await fill('#tags', `uji${run}, Form Uji`);
  await fill('#author', 'Redaksi Uji');
  await fill('#status', 'published');
  await click('#publish-btn');
  check('tulis: terbit lalu kembali ke daftar', await until(`location.pathname.startsWith('/dashboard/news') && !location.pathname.includes('create')`, 12000));
  const created = await (await api(`/news/slug/${slug}?track=0`)).json().catch(() => ({}));
  check('tulis: tersimpan dengan data yang diisi', created.title === title && created.category === 'Pendidikan' && created.author === 'Redaksi Uji' && /^\/uploads\//.test(created.imageUrl) && created.tags.includes(`uji${run}`), created);
  let p = await pageHas(`/berita/${slug}`, 'Paragraf pertama berita uji.');
  check('situs: berita baru langsung tampil', p.status === 200 && p.has, p);
  check('situs: muncul di beranda', (await pageHas('/', title)).has);
  check('situs: muncul di kategori Pendidikan', (await pageHas('/kategori/pendidikan', title)).has);
  await go(`/berita/${slug}`);
  check('situs: gambar utama termuat', await until(`(() => { const i = document.querySelector('.article-hero-img img'); return !!i && i.complete && i.naturalWidth > 0; })()`));
  check('berita: penulis & tanggal WIB tampil', (await ev(`document.querySelector('.byline-people')?.textContent.includes('Redaksi Uji') && /WIB/.test(document.querySelector('.byline-meta')?.textContent || '')`)) === true);
  check('berita: tombol bagikan berisi alamat halaman', (await ev(`(() => { const wa = document.querySelector('a[data-share="wa"]'); return !!wa && wa.href.startsWith('https://wa.me/') && wa.href.includes(encodeURIComponent(location.pathname)); })()`)) === true);
  check('berita: Berita Terkait tampil', (await ev(`document.querySelectorAll('.related-card').length`)) > 0);

  // --- sunting berita
  await go(`/dashboard/news/edit?id=${created.ID}`);
  check('sunting: isi berita termuat', await until(`document.getElementById('post-title').textContent === ${js(title)} && document.querySelectorAll('.gb-block').length > 0`));
  await ev(`(() => { const t = document.getElementById('post-title'); t.textContent = ${js(title + ' Diubah')}; t.dispatchEvent(new Event('input')); document.querySelector('.block-para').innerHTML = 'Paragraf hasil sunting.'; return true; })()`);
  await click('#publish-btn');
  check('sunting: tersimpan', await until(`location.pathname.startsWith('/dashboard/news') && !location.pathname.includes('edit')`, 12000));
  p = await pageHas(`/berita/${slug}`, 'Paragraf hasil sunting.');
  check('situs: hasil sunting langsung tampil, alamat tetap', p.status === 200 && p.has && (await pageHas(`/berita/${slug}`, title + ' Diubah')).has, p);

  // --- kategori
  await go('/dashboard/categories/');
  await until(`document.querySelectorAll('#cat-rows .btn-edit').length > 0`);
  await click('#btn-add-category');
  await fill('#cat-name', `Kategori Uji ${run}`);
  await submit('#category-form');
  check('kategori: tambah', await until(`[...document.querySelectorAll('.category-name')].some((e) => e.textContent === ${js(`Kategori Uji ${run}`)})`));
  check('kategori: slug otomatis', (await pageHas(`/kategori/kategori-uji-${run}`, `Kategori Uji ${run}`)).status === 200);
  await ev(`[...document.querySelectorAll('#cat-rows tr')].find((r) => r.textContent.includes(${js(`Kategori Uji ${run}`)})).querySelector('.btn-edit').click()`);
  await fill('#cat-name', `Kategori Ganti ${run}`);
  await fill('#cat-slug', `kategori-ganti-${run}`);
  await submit('#category-form');
  check('kategori: ubah', await until(`[...document.querySelectorAll('.category-name')].some((e) => e.textContent === ${js(`Kategori Ganti ${run}`)})`));
  await ev(`[...document.querySelectorAll('#cat-rows tr')].find((r) => r.textContent.includes(${js(`Kategori Ganti ${run}`)})).querySelector('.btn-delete').click()`);
  check('kategori: hapus', await until(`![...document.querySelectorAll('.category-name')].some((e) => e.textContent.includes(${js(run)}))`));

  // --- poster
  await go('/dashboard/banners/');
  await until(`!document.querySelector('#banner-rows')?.textContent.includes('Memuat')`);
  await click('#btn-add-banner');
  await upload('#file-input');
  check('poster: unggah foto', await until(`document.getElementById('preview-img').src.includes('/uploads/')`));
  await fill('#banner-title', `Poster Uji ${run}`);
  await fill('#banner-target-url', 'https://contoh.id/acara');
  await submit('#banner-form');
  const bannerRow = `[...document.querySelectorAll('#banner-rows tr')].find((r) => r.textContent.includes(${js(`Poster Uji ${run}`)}))`;
  check('poster: simpan', await until(`!!${bannerRow} && ${bannerRow}.textContent.includes('Aktif di Beranda')`));
  await ev(`${bannerRow}.querySelector('.btn-toggle').click()`);
  check('poster: nonaktifkan', await until(`${bannerRow}?.textContent.includes('Nonaktif')`));
  await ev(`${bannerRow}.querySelector('.btn-delete').click()`);
  check('poster: hapus', await until(`!${bannerRow}`));

  // --- pengaturan situs
  await go('/dashboard/settings/');
  await until(`document.getElementById('site_description') !== null`);
  await sleep(500);
  await fill('#site_description', `Deskripsi situs uji ${run}`);
  await fill('#contact_email', 'redaksi@contoh.id');
  await fill('#instagram_url', 'https://instagram.com/uji');
  await submit('#settings-form');
  check('pengaturan: simpan', await until(`!document.getElementById('settings-ok').hidden`));
  check('situs: deskripsi & kontak di footer', await eventually('/', `Deskripsi situs uji ${run}`) && (await pageHas('/', 'redaksi@contoh.id')).has && (await pageHas('/', 'https://instagram.com/uji')).has);
  await fill('#instagram_url', 'bukan tautan');
  await submit('#settings-form');
  check('pengaturan: tautan salah ditolak', await until(`!document.getElementById('settings-error').hidden`) || (await ev(`document.getElementById('instagram_url').validity.valid`)) === false);

  // --- halaman informasi
  await go('/dashboard/pages/');
  await until(`!document.getElementById('page-rows').textContent.includes('Memuat')`);
  await click('#btn-add-page');
  await fill('#pg-title', `Tentang Kami ${run}`);
  await fill('#pg-order', '1');
  await fill('#pg-content', `Paragraf pertama halaman.\n\n## Redaksi\nEmail: redaksi@contoh.id\n<script>window.__xss=1</script>`);
  await submit('#page-form');
  const pageSlug = `tentang-kami-${run}`;
  check('halaman: simpan', await until(`[...document.querySelectorAll('#page-rows tr')].some((r) => r.dataset.title === ${js(`Tentang Kami ${run}`)})`));
  p = await pageHas(`/halaman/${pageSlug}`, 'Paragraf pertama halaman.');
  check('situs: halaman informasi tampil', p.status === 200 && p.has, p);
  check('situs: halaman di bar atas & footer', await eventually('/', `href="/halaman/${pageSlug}"`));
  await go(`/halaman/${pageSlug}`);
  check('situs: isi halaman aman (script tidak jalan) & tautan email', (await ev(`window.__xss === undefined && !!document.querySelector('.info-content a[href="mailto:redaksi@contoh.id"]') && !!document.querySelector('.info-content h2')`)) === true);

  // --- poster di posisi tengah beranda
  await go('/dashboard/banners/');
  await until(`!document.querySelector('#banner-rows')?.textContent.includes('Memuat')`);
  await click('#btn-add-banner');
  await upload('#file-input');
  await until(`document.getElementById('preview-img').src.includes('/uploads/')`);
  await fill('#banner-title', `Poster Tengah ${run}`);
  await fill('#banner-placement', 'tengah');
  await submit('#banner-form');
  check('poster: simpan di posisi tengah', await until(`[...document.querySelectorAll('#banner-rows tr')].some((r) => r.textContent.includes(${js(`Poster Tengah ${run}`)}))`));
  check('situs: poster tampil di beranda', await eventually('/', `title="Poster Tengah ${run}"`));
  const middleRow = `[...document.querySelectorAll('#banner-rows tr')].find((r) => r.textContent.includes(${js(`Poster Tengah ${run}`)}))`;
  await ev(`(() => { const s = ${middleRow}.querySelector('.placement-select'); s.value = 'header'; s.dispatchEvent(new Event('change')); return true; })()`);
  check('poster: pindah ke header', await eventually('/kategori/semua', `title="Poster Tengah ${run}"`));
  await ev(`${middleRow}.querySelector('.btn-delete').click()`);
  check('poster: hapus → hilang dari situs', await until(`!${middleRow}`) && await eventually('/', `title="Poster Tengah ${run}"`, false));

  // --- kategori ke menu navigasi & topik pilihan
  await go('/dashboard/categories/');
  await until(`document.querySelectorAll('#cat-rows .btn-edit').length > 0`);
  await click('#btn-add-category');
  await fill('#cat-name', `Menu Uji ${run}`);
  await fill('#cat-menu', '1');
  await ev(`document.getElementById('cat-topic').checked = true`);
  await submit('#category-form');
  check('kategori: tambah ke menu', await until(`[...document.querySelectorAll('.category-name')].some((e) => e.textContent === ${js(`Menu Uji ${run}`)})`));
  check('situs: kategori muncul di menu navigasi', await eventually('/', `<a href="/kategori/menu-uji-${run}"`));
  check('situs: kategori muncul di Topik Pilihan', (await pageHas('/', `topic-card`)).has && (await pageHas('/', `Menu Uji ${run}</strong>`)).has);
  await ev(`[...document.querySelectorAll('#cat-rows tr')].find((r) => r.textContent.includes(${js(`Menu Uji ${run}`)})).querySelector('.btn-delete').click()`);
  check('kategori: hapus → hilang dari menu', await eventually('/', `/kategori/menu-uji-${run}"`, false));

  // --- event berbayar
  await go('/dashboard/events/');
  await until(`!document.getElementById('event-rows').textContent.includes('Memuat')`);
  await fill('#bank-name', 'BCA');
  await fill('#bank-number', `77${run}`);
  await fill('#bank-holder', 'Edukator Uji');
  await submit('#bank-form');
  check('rekening: tambah', await until(`document.getElementById('bank-rows').textContent.includes(${js(`77${run}`)})`));
  await click('#btn-add-event');
  await fill('#ev-title', `Webinar Uji ${run}`);
  await fill('#ev-venue', 'Zoom Meeting');
  await fill('#ev-start', '2031-02-01T09:00');
  await fill('#ev-end', '2031-02-01T11:00');
  await fill('#ev-price', '15000');
  await fill('#ev-desc', 'Baris satu acara.\nBaris dua acara.');
  await upload('#ev-image-file');
  check('event: unggah gambar', await until(`document.getElementById('ev-image').value.includes('/uploads/')`));
  await click('#add-speaker');
  await ev(`(() => { const r = document.querySelector('#speaker-list .repeat-row'); r.querySelector('.sp-name').value = 'Dr. Pemateri Uji'; r.querySelector('.sp-role').value = 'Dosen'; return true; })()`);
  await ev(`(() => { [...document.querySelectorAll('#bank-checks input')].forEach((b) => { if (b.parentElement.textContent.includes(${js(`77${run}`)})) b.checked = true; }); return true; })()`);
  await submit('#event-form');
  eventRow = `[...document.querySelectorAll('#event-rows tr')].find((r) => r.dataset.title === ${js(`Webinar Uji ${run}`)})`;
  check('event: simpan', await until(`!!${eventRow}`));
  const eventSlug = `webinar-uji-${run}`;
  p = await pageHas(`/event/${eventSlug}`, 'Dr. Pemateri Uji');
  check('situs: event tampil dengan pemateri', p.status === 200 && p.has, p);
  check('situs: jam event WIB benar', (await pageHas(`/event/${eventSlug}`, '09.00')).has || (await pageHas(`/event/${eventSlug}`, '09:00')).has);
  check('situs: event di daftar Event & Webinar', (await pageHas('/event', `Webinar Uji ${run}`)).has);
  check('situs: event di Agenda beranda', await eventually('/', `Webinar Uji ${run}`));
  await ev(`${eventRow}.querySelector('.ev-edit').click()`);
  check('event: form ubah terisi (WIB)', await until(`document.getElementById('ev-title').value === ${js(`Webinar Uji ${run}`)} && document.getElementById('ev-start').value === '2031-02-01T09:00' && document.querySelectorAll('#speaker-list .repeat-row').length === 1`));
  await fill('#ev-venue', 'Zoom Meeting (diubah)');
  await submit('#event-form');
  check('event: ubah', await until(`!document.getElementById('event-modal').classList.contains('active')`) && (await pageHas(`/event/${eventSlug}`, 'Zoom Meeting (diubah)')).has);

  // --- newsletter & pencarian
  await go('/');
  await fill('#emailInput', 'a@b');
  await submit('#newsletterForm');
  check('newsletter: email tidak valid ditolak', await until(`document.getElementById('newsletterMessage').textContent.includes('tidak valid')`));
  const subscriber = `pembaca-${run}@contoh.id`;
  await fill('#emailInput', subscriber);
  await submit('#newsletterForm');
  check('newsletter: daftar', await until(`document.getElementById('newsletterMessage').textContent.includes('Terima kasih')`));
  await click('#searchToggle');
  await fill('#searchInput', `Uji Form ${run}`);
  await ev(`document.getElementById('searchInput').form.requestSubmit()`);
  check('pencarian: menemukan berita baru', await until(`location.pathname.startsWith('/cari') && [...document.querySelectorAll('#searchResults .card-title')].some((e) => e.textContent.includes(${js(run)}))`));
  await go(`/cari?tag=uji${run}`);
  check('pencarian tag', await until(`document.querySelectorAll('#searchResults .news-card').length === 1`));

  await go('/dashboard/');
  check('dashboard: pelanggan newsletter tercatat', await until(`Number(document.getElementById('subscribers').textContent) >= 1`));
  check('dashboard: daftar pelanggan memuat email', (await ev(`fetch('/api/admin/newsletter', { credentials: 'include' }).then((r) => r.json()).then((l) => l.some((x) => x.email === ${js(subscriber)}))`)) === true);
  // --- akun redaksi (penulis)
  const writer = { name: `Penulis Uji ${run}`, email: `penulis-${run}@contoh.id`, password: 'rahasia-penulis-1' };
  await go('/dashboard/users/');
  await until(`!document.getElementById('user-rows').textContent.includes('Memuat')`);
  await click('#btn-add-user');
  await fill('#u-name', writer.name);
  await fill('#u-email', writer.email);
  await fill('#u-role', 'penulis');
  await fill('#u-password', writer.password);
  await submit('#user-form');
  const writerRow = `[...document.querySelectorAll('#user-rows tr')].find((r) => r.dataset.email === ${js(writer.email)})`;
  check('akun redaksi: tambah penulis', await until(`!!${writerRow} && ${writerRow}.textContent.includes('Penulis')`));

  // --- video
  await go('/dashboard/videos/');
  await until(`!document.getElementById('video-rows').textContent.includes('Memuat')`);
  await click('#btn-add-video');
  await fill('#v-url', 'https://www.youtube.com/watch?v=aqz-KE-bpKQ');
  check('video: pratinjau gambar dari alamat', await until(`!document.getElementById('v-preview').hidden && document.getElementById('v-preview').src.includes('aqz-KE-bpKQ')`));
  await fill('#v-title', `EnhaPodcast Uji ${run}`);
  await submit('#video-form');
  const videoRow = `[...document.querySelectorAll('#video-rows tr')].find((r) => r.dataset.title === ${js(`EnhaPodcast Uji ${run}`)})`;
  check('video: simpan', await until(`!!${videoRow}`));
  await fill('#video_section_title', `EnhaPodcast ${run}`);
  await submit('#video-settings');
  check('video: judul kolom tersimpan', await until(`!document.getElementById('vs-ok').hidden`));
  check('situs: kolom video di beranda', await eventually('/', `data-play="aqz-KE-bpKQ"`) && (await pageHas('/', `EnhaPodcast ${run}`)).has);
  check('situs: menu Video & halaman /video', (await pageHas('/', 'href="/video"')).has && (await pageHas('/video', `EnhaPodcast Uji ${run}`)).has);
  await go('/');
  await ev(`document.querySelector('[data-play]').click()`);
  check('situs: klik video memutar YouTube', await until(`!!document.querySelector('.video-player iframe[src*="youtube-nocookie.com/embed/aqz-KE-bpKQ"]')`));
  await go('/dashboard/'); // logout lewat tombol di dashboard

  check('logout admin', await logout());
  await go('/dashboard/');
  check('dashboard tanpa login dialihkan ke login', await until(`location.pathname.startsWith('/login')`));

  // ================= PENULIS =================
  check('penulis: login masuk ke dashboard', await login(writer.email, writer.password) && (await path_()).startsWith('/dashboard'));
  check('penulis: menu sesuai peran', await until(`document.body.dataset.role === 'penulis'`) && (await ev(`(() => { const vis = (h) => { const a = document.querySelector('.sidebar-nav a[href="' + h + '"]'); return !!a && !a.hidden; }; return vis('/dashboard/news') && vis('/dashboard/activity') && !vis('/dashboard/users') && !vis('/dashboard/categories') && !vis('/dashboard/videos') && !vis('/dashboard/settings'); })()`)) === true);
  await go('/dashboard/users/');
  check('penulis: halaman akun redaksi dialihkan', await until(`location.pathname === '/dashboard' || location.pathname === '/dashboard/'`));
  const writerTitle = `Tulisan Penulis Uji ${run}`;
  await go('/dashboard/news/create/');
  await until(`document.getElementById('category').options.length > 2`);
  check('penulis: isian nama penulis & unggulan disembunyikan', (await ev(`getComputedStyle(document.getElementById('author').closest('.sb-field')).display === 'none' && getComputedStyle(document.getElementById('is-featured').closest('.sb-section')).display === 'none'`)) === true);
  await ev(`(() => { const t = document.getElementById('post-title'); t.textContent = ${js(writerTitle)}; t.dispatchEvent(new Event('input')); document.querySelector('.block-para').innerHTML = 'Isi tulisan penulis uji.'; return true; })()`);
  await fill('#category', 'Pendidikan');
  await fill('#status', 'published');
  await click('#publish-btn');
  check('penulis: terbitkan berita', await until(`location.pathname.startsWith('/dashboard/news') && !location.pathname.includes('create')`, 12000));
  const writerNews = await (await api(`/news/slug/${slug.replace(/^berita-uji-form/, 'tulisan-penulis-uji')}?track=0`)).json().catch(() => ({}));
  check('penulis: nama di berita = nama akun', writerNews.author === writer.name && writerNews.createdByName === writer.name, writerNews);
  await go('/dashboard/news/');
  check('penulis: daftar berita hanya tulisannya', await until(`document.querySelectorAll('.news-row').length === 1 && document.querySelector('.news-row .td-title').textContent === ${js(writerTitle)}`));
  await go(`/dashboard/news/edit?id=${created.ID}`);
  check('penulis: berita orang lain hanya baca', await until(`document.getElementById('save-indicator')?.textContent.includes('Hanya baca') || document.body.textContent.includes('Anda hanya dapat mengubah tulisan Anda sendiri')`));
  await go(`/dashboard/news/edit?id=${writerNews.ID}`);
  check('penulis: riwayat perubahan tampil di halaman edit', await until(`document.getElementById('history-list').textContent.includes(${js(writer.name)}) && document.getElementById('history-who').textContent.includes('Dibuat oleh akun')`));
  check('logout penulis', await logout());

  // ================= PENGGUNA =================
  const userEmail = `pembaca-${run}@contoh.id`;
  await go('/register');
  await fill('#name', 'Pembaca Uji');
  await fill('#email', userEmail);
  await fill('#password', 'rahasia-awal-1');
  await fill('#confirm-password', 'rahasia-awal-1');
  await ev(`document.getElementById('terms').checked = true`);
  await submit('#register-form');
  check('daftar akun lewat form', await until(`location.pathname.startsWith('/user')`, 12000));
  check('profil pengguna termuat', await until(`document.getElementById('p-name')?.textContent === 'Pembaca Uji'`));

  await go(`/berita/${slug}`);
  check('komentar: form muncul setelah login', await until(`getComputedStyle(document.getElementById('main-comment-form')).display !== 'none'`));
  await fill('#comment-content', 'Artikel yang sangat bermanfaat, terima kasih.');
  await submit('#comment-form');
  check('komentar: terkirim dan tampil', await until(`document.getElementById('comments-list').textContent.includes('sangat bermanfaat')`));
  await until(`!!document.querySelector('.btn-like')`);
  await click('.btn-like');
  check('komentar: suka', await until(`document.querySelector('.btn-like').textContent.trim() === '1'`, 6000) || (await api(`/comments/news/${created.ID}`).then((r) => r.json())).some((c) => c.like_count === 1));
  check('simpan berita: tombol muncul', await until(`!document.getElementById('bookmark-btn').hidden`));
  await click('#bookmark-btn');
  check('simpan berita', await until(`document.getElementById('bm-text').textContent === 'Tersimpan'`));
  // --- daftar event berbayar dengan bukti transfer
  await go(`/event/webinar-uji-${run}`);
  check('event: form pendaftaran terisi dari akun', await until(`document.getElementById('reg-email')?.value === ${js(userEmail)}`));
  await fill('#reg-phone', '081234567890');
  await upload('#reg-proof');
  check('event: unggah bukti transfer', await until(`document.getElementById('proof-url').value.includes('/uploads/')`));
  check('event: rekening tujuan tampil', (await ev(`document.body.textContent.includes(${js(`77${run}`)})`)) === true);
  await submit('#reg-form');
  check('event: pendaftaran terkirim', await until(`getComputedStyle(document.getElementById('success-modal')).display !== 'none' && /^TIC-/.test(document.getElementById('modal-ticket-code').textContent)`));
  await go('/user/tickets/');
  check('Tiket Saya: menunggu verifikasi', await until(`document.getElementById('tickets-list').textContent.includes(${js(`Webinar Uji ${run}`)}) && !!document.querySelector('.ticket-card .badge-waiting')`));
  await go('/user/bookmarks/');
  check('halaman Berita Disimpan memuat berita', await until(`document.getElementById('bookmark-list').textContent.includes(${js(run)})`));
  await go('/user/history/');
  check('Riwayat Baca memuat berita', await until(`document.getElementById('history-list').textContent.includes(${js(run)})`));
  await go('/user/inbox/');
  check('Kotak Masuk memuat pesan selamat datang', await until(`document.getElementById('notif-list').textContent.includes('Selamat datang')`));
  await until(`getComputedStyle(document.getElementById('mark-all-read')).display !== 'none'`, 3000);
  await click('#mark-all-read');
  check('tandai semua dibaca', await until(`document.querySelectorAll('.notif-card.unread').length === 0`));

  await go('/user/edit/');
  await until(`document.getElementById('name').value === 'Pembaca Uji'`);
  await upload('#avatar-input');
  check('profil: unggah foto', await until(`document.getElementById('avatar-url').value.includes('/uploads/')`));
  await fill('#name', 'Pembaca Diubah');
  await fill('#bio', 'Suka membaca berita pendidikan.');
  await submit('#edit-profile-form');
  check('profil: simpan', await until(`location.pathname.startsWith('/user') && document.getElementById('p-name')?.textContent === 'Pembaca Diubah'`, 12000));
  check('profil: bio & foto tampil', await until(`document.getElementById('p-bio').textContent.includes('Suka membaca') && document.getElementById('p-avatar').naturalWidth > 0`));

  await go('/user/password/');
  await fill('#old-password', 'rahasia-awal-1');
  await fill('#new-password', 'rahasia-baru-2');
  await fill('#confirm-password', 'beda');
  await submit('#password-form');
  check('ganti password: ulangan beda ditolak', await until(`document.getElementById('error-msg').textContent.includes('tidak sama')`));
  await fill('#old-password', 'salah-total');
  await fill('#confirm-password', 'rahasia-baru-2');
  await submit('#password-form');
  check('ganti password: password lama salah ditolak', await until(`document.getElementById('error-msg').textContent.toLowerCase().includes('salah')`));
  await fill('#old-password', 'rahasia-awal-1');
  await submit('#password-form');
  check('ganti password: berhasil', await until(`document.getElementById('success-msg').textContent.includes('berhasil')`));
  check('logout pengguna', await logout());
  check('login dengan password baru', await login(userEmail, 'rahasia-baru-2') && (await path_()).startsWith('/user'));
  await logout();

  // ================= ADMIN: bersih-bersih =================
  await login(ADMIN.email, ADMIN.password);
  await go('/dashboard/registrations/');
  const regRow = `[...document.querySelectorAll('#reg-rows tr')].find((r) => r.dataset.email === ${js(userEmail)})`;
  check('pendaftaran: tampil dengan bukti', await until(`!!${regRow} && !!${regRow}.querySelector('a[href*="/uploads/"]')`));
  await ev(`${regRow}.querySelector('.reg-verify').click()`);
  check('pendaftaran: verifikasi', await until(`${regRow}?.textContent.includes('Terverifikasi')`));
  check('peserta: tiket terverifikasi', (await ev(`fetch('/api/admin/registrations', { credentials: 'include' }).then((r) => r.json()).then((d) => d.data.some((x) => x.email === ${js(userEmail)} && x.status === 'verified'))`)) === true);
  await go('/dashboard/events/');
  await until(`!!${eventRow}`);
  await ev(`${eventRow}.querySelector('.ev-delete').click()`);
  check('event: hapus → hilang dari situs', await until(`!${eventRow}`) && (await pageHas(`/event/webinar-uji-${run}`, 'x')).status === 404);
  await ev(`[...document.querySelectorAll('#bank-rows tr')].find((r) => r.textContent.includes(${js(`77${run}`)})).querySelector('button').click()`);
  check('rekening: hapus', await until(`!document.getElementById('bank-rows').textContent.includes(${js(`77${run}`)})`));
  await go('/dashboard/pages/');
  await until(`[...document.querySelectorAll('#page-rows tr')].some((r) => r.dataset.title === ${js(`Tentang Kami ${run}`)})`);
  await ev(`[...document.querySelectorAll('#page-rows tr')].find((r) => r.dataset.title === ${js(`Tentang Kami ${run}`)}).querySelector('.pg-delete').click()`);
  check('halaman: hapus → 404', await until(`![...document.querySelectorAll('#page-rows tr')].some((r) => r.dataset.title === ${js(`Tentang Kami ${run}`)})`) && (await pageHas(`/halaman/tentang-kami-${run}`, 'x')).status === 404);

  await go('/dashboard/activity/');
  check('riwayat: aktivitas penulis tercatat', await until(`[...document.querySelectorAll('#activity-rows tr')].some((r) => r.textContent.includes(${js(writer.name)}) && r.textContent.includes('Membuat') && r.textContent.includes(${js(`Tulisan Penulis Uji ${run}`)}))`));
  check('hapus berita penulis', (await ev(`fetch('/api/news/' + ${js(String((await (await api(`/news?q=${encodeURIComponent(`Tulisan Penulis Uji ${run}`)}`)).json())[0]?.ID))}, { method: 'DELETE', credentials: 'include' }).then((r) => r.ok)`)) === true);
  await go('/dashboard/users/');
  const writerRow2 = `[...document.querySelectorAll('#user-rows tr')].find((r) => r.dataset.email === ${js(`penulis-${run}@contoh.id`)})`;
  await until(`!!${writerRow2}`);
  await ev(`${writerRow2}.querySelector('.u-edit').click()`);
  await ev(`document.getElementById('u-suspended').checked = true`);
  await submit('#user-form');
  check('akun redaksi: nonaktifkan penulis', await until(`${writerRow2}?.textContent.includes('Nonaktif')`));
  check('akun nonaktif tidak bisa masuk', (await api('/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email: `penulis-${run}@contoh.id`, password: 'rahasia-penulis-1' }) })).status === 403);
  await go('/dashboard/videos/');
  const videoRow2 = `[...document.querySelectorAll('#video-rows tr')].find((r) => r.dataset.title === ${js(`EnhaPodcast Uji ${run}`)})`;
  await until(`!!${videoRow2}`);
  await ev(`${videoRow2}.querySelector('.v-delete').click()`);
  check('video: hapus → hilang dari beranda', await until(`!${videoRow2}`) && await eventually('/', `EnhaPodcast Uji ${run}`, false));
  await fill('#video_section_title', '');
  await submit('#video-settings');

  await go('/dashboard/comments/');
  check('analisis komentar memuat komentar', await until(`document.getElementById('comments-list').textContent.includes('sangat bermanfaat')`));
  await ev(`[...document.querySelectorAll('.comment-item')].find((c) => c.textContent.includes('sangat bermanfaat')).querySelector('.btn-danger').click()`);
  check('hapus komentar', await until(`!document.getElementById('comments-list').textContent.includes('sangat bermanfaat')`));

  await go('/dashboard/news/');
  await fill('#search-input', `Uji Form ${run}`);
  check('daftar berita: cari', await until(`document.querySelectorAll('.news-row').length === 1`));
  await click('.action-delete');
  await click('#confirm-delete-btn');
  check('hapus berita', await until(`document.querySelectorAll('.news-row').length === 0`));
  p = await pageHas(`/berita/${slug}`, title);
  check('situs: berita terhapus langsung hilang (404)', p.status === 404, p.status);
  check('situs: hilang dari beranda', !(await pageHas('/', title)).has);
  check('situs: hilang dari kategori', !(await pageHas('/kategori/pendidikan', title)).has);
  check('situs: hilang dari pencarian', (await (await api(`/news?q=${encodeURIComponent('Uji Form ' + run)}`)).json()).length === 0);
  await logout();
} catch (e) {
  failures.push(`berhenti: ${e.stack || e}`);
}

check('tanpa error JavaScript di halaman', jsErrors.length === 0, jsErrors.slice(0, 5));
ws.close();
chrome.kill();
console.log(`\nFORM LULUS ${pass}, GAGAL ${failures.length}`);
failures.forEach((f) => console.log('  ✗ ' + f));
if (process.env.GITHUB_ACTIONS) console.log(`::${failures.length ? 'error' : 'notice'}::Form lulus ${pass}, gagal ${failures.length}%0A${failures.join('%0A')}`);
process.exit(failures.length ? 1 : 0);

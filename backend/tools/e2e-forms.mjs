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
  await until(`!!document.getElementById('logout-btn')`);
  await click('#logout-btn');
  return until(`location.pathname === '/'`);
}

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
  check('logout admin', await logout());
  await go('/dashboard/');
  check('dashboard tanpa login dialihkan ke login', await until(`location.pathname.startsWith('/login')`));

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

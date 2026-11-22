// Uji menyeluruh API backend yang sedang berjalan (Node 22+, tanpa dependensi).
//
//   node tools/smoke-test.mjs
//   BASE_URL=http://localhost:3000 ADMIN_EMAIL=... ADMIN_PASSWORD=... node tools/smoke-test.mjs
//
// Membuat lalu menghapus data uji (hapus di sistem ini = soft delete, jadi barisnya tetap ada
// di database). Karena itu hanya boleh ke localhost; jangan dipakai pada database produksi.

const BASE = process.env.BASE_URL || 'http://localhost:3000';
if (!/^https?:\/\/(localhost|127\.0\.0\.1)(:|\/|$)/.test(BASE)) {
  console.error('Ditolak: skrip ini menulis data uji dan hanya boleh dijalankan ke localhost.');
  process.exit(2);
}
const ADMIN = { email: process.env.ADMIN_EMAIL || 'admin@edutech.id', password: process.env.ADMIN_PASSWORD || 'admin123' };
const run = Date.now();

let pass = 0;
const failures = [];
function check(name, ok, info) {
  if (ok) pass++;
  else failures.push(`${name}${info === undefined ? '' : ' → ' + JSON.stringify(info).slice(0, 300)}`);
  console.log(`${ok ? '✓' : '✗'} ${name}`);
}

async function call(method, path, { body, cookie, form, headers = {} } = {}) {
  const res = await fetch(BASE + path, {
    method,
    headers: { ...(body ? { 'Content-Type': 'application/json' } : {}), ...(cookie ? { Cookie: cookie } : {}), ...headers },
    body: form || (body ? JSON.stringify(body) : undefined),
  });
  const text = await res.text();
  let data = null;
  try { data = JSON.parse(text); } catch { data = text; }
  return { status: res.status, data, headers: res.headers };
}
const cookieOf = (r) => (r.headers.get('set-cookie') || '').split(';')[0];

// ---------- masuk & hak akses ----------
check('login password salah ditolak', (await call('POST', '/api/auth/login', { body: { ...ADMIN, password: 'salah-sekali' } })).status === 401);
const login = await call('POST', '/api/auth/login', { body: ADMIN });
const admin = cookieOf(login);
check('login admin', login.status === 200 && admin.startsWith('auth_token='), login.data);
check('cookie login HttpOnly', /httponly/i.test(login.headers.get('set-cookie') || ''));
const me = await call('GET', '/api/auth/profile', { cookie: admin });
check('profil admin (dipakai layout dashboard)', me.status === 200 && me.data.role === 'admin' && me.data.id > 0, me.data);
check('profil tanpa login ditolak', (await call('GET', '/api/auth/profile')).status === 401);

const email = `uji-${run}@contoh.id`;
const reg = await call('POST', '/api/auth/register', { body: { name: 'Pengguna Uji', email, password: 'rahasia123' } });
const user = cookieOf(reg);
check('daftar pengguna', reg.status === 200 && user.startsWith('auth_token='), reg.data);
check('email ganda ditolak', (await call('POST', '/api/auth/register', { body: { name: 'Ganda', email, password: 'rahasia123' } })).status >= 400);
for (const [m, p] of [['GET', '/api/admin/stats'], ['GET', '/api/admin/banners'], ['GET', '/api/comments'], ['POST', '/api/news'], ['POST', '/api/categories']]) {
  check(`${m} ${p}: tanpa login 401`, (await call(m, p, { body: m === 'POST' ? {} : undefined })).status === 401);
  check(`${m} ${p}: pengguna biasa 403`, (await call(m, p, { cookie: user, body: m === 'POST' ? {} : undefined })).status === 403);
}
check('token palsu ditolak', (await call('GET', '/api/auth/profile', { cookie: 'auth_token=eyJhbGciOiJub25lIn0.eyJ1c2VySWQiOjEsInJvbGUiOiJhZG1pbiJ9.' })).status === 401);

// ---------- kategori ----------
const catName = `Uji & Coba ${run}`;
const cat = await call('POST', '/api/categories', { cookie: admin, body: { name: catName } });
check('buat kategori, slug otomatis bersih', cat.status === 200 && cat.data.slug === `uji-coba-${run}`, cat.data);
const cat2 = await call('POST', '/api/categories', { cookie: admin, body: { name: `Kedua ${run}` } });
check('buat kategori kedua', cat2.status === 200 && cat2.data.ID > 0, cat2.data);
check('daftar kategori', (await call('GET', '/api/categories')).data.some((c) => c.name === catName));

// ---------- unggah gambar ----------
const png = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==', 'base64');
const upload = (cookie, name = 'uji.png', bytes = png) => {
  const form = new FormData();
  form.append('image', new Blob([bytes]), name);
  return call('POST', '/api/upload', { cookie, form });
};
const up = await upload(admin);
check('unggah gambar', up.status === 200 && /^\/uploads\//.test(up.data.url || ''), up.data);
const img = await fetch(BASE + up.data.url);
check('gambar unggahan bisa dibuka', img.status === 200 && /^image/.test(img.headers.get('content-type') || ''));
check('unggah tanpa login ditolak', (await upload()).status === 401);
check('unggah bukan gambar ditolak', (await upload(admin, 'x.exe', Buffer.from('MZ'))).status === 400);
check('unggah gambar palsu ditolak', (await upload(admin, 'x.png', Buffer.from('bukan gambar'))).status === 400);

// ---------- berita ----------
const title = `Berita Uji ${run}`;
const make = (extra) => call('POST', '/api/news', { cookie: admin, body: { title, content: `<p>Isi&nbsp;uji &amp; coba ${'kata '.repeat(450)}</p>`, category: catName, tags: `uji${run}, Kedua Kata`, imageUrl: up.data.url, status: 'published', ...extra } });
const n = await make();
const art = n.data.data || {};
check('terbitkan berita', n.status === 201 && art.ID > 0 && art.slug === `berita-uji-${run}`, n.data);
check('slug kategori terisi', art.categorySlug === `uji-coba-${run}`, art.categorySlug);
check('waktu baca dihitung', art.readTime === 3, art.readTime);
check('ringkasan tanpa entitas HTML', /Isi uji & coba/.test(art.excerpt) && !/&nbsp;|&amp;/.test(art.excerpt), art.excerpt);
const twin = await make();
check('judul sama → slug unik', twin.status === 201 && twin.data.data.slug !== art.slug, twin.data?.data?.slug);
const draft = await make({ title: `Draft Uji ${run}`, status: 'draft' });
const draftArt = draft.data.data || {};
check('simpan draft', draft.status === 201 && draftArt.isPublished === false, draftArt.isPublished);
check('draft tersembunyi dari publik (slug)', (await call('GET', `/api/news/slug/${draftArt.slug}`)).status === 404);
check('draft tersembunyi dari publik (id)', (await call('GET', `/api/news/${draftArt.ID}`)).status === 404);
check('draft terlihat admin', (await call('GET', `/api/news/${draftArt.ID}`, { cookie: admin })).status === 200);
check('judul kosong ditolak', (await make({ title: ' ' })).status === 400);

const q = (s) => call('GET', `/api/news?${s}`);
let r = await q('limit=100');
check('daftar berita publik tanpa draft', r.status === 200 && r.data.some((a) => a.ID === art.ID) && !r.data.some((a) => a.ID === draftArt.ID));
check('header X-Total-Count', Number(r.headers.get('x-total-count')) >= 2, r.headers.get('x-total-count'));
check('daftar memuat slug kategori', r.data.find((a) => a.ID === art.ID)?.categories?.includes(`uji-coba-${run}`));
r = await q('limit=1&lite=1');
check('lite tanpa konten', r.data.length === 1 && r.data[0].content === '');
const [p1, p2] = [await q('limit=1&offset=0'), await q('limit=1&offset=1')];
check('paginasi offset', p1.data[0].ID !== p2.data[0].ID);
check('limit dibatasi 100', (await q('limit=5000')).data.length <= 100);
for (const [name, s] of [['slug kategori', `category=uji-coba-${run}`], ['nama kategori', `category=${encodeURIComponent(catName)}`], ['tag', `tag=uji${run}`], ['tag berspasi', 'tag=Kedua%20Kata'], ['kata kunci', `q=${encodeURIComponent(title)}`]]) {
  r = await q(s);
  check(`filter ${name}`, r.status === 200 && r.data.some((a) => a.ID === art.ID), r.data?.length);
}
check('route kategori', (await call('GET', `/api/news/category/uji-coba-${run}`)).data.some((a) => a.ID === art.ID));
check('kata kunci tak ada → array kosong', JSON.stringify((await q('q=zzzqqq-tidak-ada')).data) === '[]');
check('kata kunci karakter khusus', (await q(`q=${encodeURIComponent(`%' OR 1=1 -- "`)}`)).status === 200);
check('publik minta status=draft → tetap terbit saja', !(await q('status=draft&limit=100')).data.some((a) => a.ID === draftArt.ID));
check('publik minta status=all → tetap terbit saja', !(await q('status=all&limit=100')).data.some((a) => a.ID === draftArt.ID));
r = await call('GET', '/api/news?status=all&limit=100', { cookie: admin });
check('admin status=all termasuk draft', r.data.some((a) => a.ID === draftArt.ID));
r = await call('GET', '/api/news?status=draft&limit=100', { cookie: admin });
check('admin status=draft hanya draft', r.data.length > 0 && r.data.every((a) => a.status === 'draft'));
for (const p of ['featured', 'trending', 'popular?days=7', `by-categories?cats=uji-coba-${run}`]) check(`GET /api/news/${p.split('?')[0]}`, (await call('GET', `/api/news/${p}`)).status === 200);
check('trending berisi (fallback terbanyak dibaca)', (await call('GET', '/api/news/trending')).data.length > 0);

const views = async () => (await call('GET', `/api/news/${art.ID}`)).data.viewCount;
const v0 = await views();
check('hitung tayangan', (await call('POST', `/api/news/${art.ID}/view`)).status === 200 && (await views()) === v0 + 1);
check('tayangan draft ditolak', (await call('POST', `/api/news/${draftArt.ID}/view`)).status === 404);
check('tayangan id ngawur ditolak', (await call('POST', '/api/news/abc/view')).status === 404);

// ---------- sunting ----------
r = await call('PUT', `/api/news/${art.ID}`, { cookie: admin, body: { title: `${title} Baru`, content: '<p>Pendek &amp; padat</p>', slug: `slug-baru-${run}`, category: `Kedua ${run}`, tags: 'a, b', status: 'published', imageUrl: up.data.url, isFeatured: true, isTrending: false } });
check('sunting berita', r.status === 200, r.data);
r = await call('GET', `/api/news/slug/slug-baru-${run}`);
check('slug baru aktif', r.status === 200 && r.data.title === `${title} Baru`, r.status);
check('waktu baca & ringkasan ikut berubah', r.data.readTime === 1 && r.data.excerpt === 'Pendek & padat', [r.data.readTime, r.data.excerpt]);
r = await call('GET', `/api/news/resolve/${art.slug}`);
check('slug lama dialihkan', r.status === 200 && r.data.slug === `slug-baru-${run}`, r.data);
check('slug tak dikenal 404', (await call('GET', '/api/news/resolve/tidak-ada-sama-sekali')).status === 404);
check('pindah kategori: muncul di yang baru', (await q(`category=kedua-${run}`)).data.some((a) => a.ID === art.ID));
check('pindah kategori: hilang dari yang lama', !(await q(`category=uji-coba-${run}`)).data.some((a) => a.ID === art.ID));
check('tanda unggulan', (await call('GET', '/api/news/featured?limit=20')).data.some((a) => a.ID === art.ID));
check('slug bentrok ditolak', (await call('PUT', `/api/news/${art.ID}`, { cookie: admin, body: { slug: twin.data.data.slug } })).status === 400);
check('terbit → draft', (await call('PUT', `/api/news/${art.ID}`, { cookie: admin, body: { status: 'draft' } })).status === 200 && (await call('GET', `/api/news/${art.ID}`)).status === 404);
check('draft → terbit', (await call('PUT', `/api/news/${art.ID}`, { cookie: admin, body: { status: 'published' } })).status === 200 && (await call('GET', `/api/news/${art.ID}`)).status === 200);

r = await call('PUT', `/api/categories/${cat2.data.ID}`, { cookie: admin, body: { name: `Ganti ${run}`, slug: `ganti-${run}` } });
check('ganti nama kategori', r.status === 200 && r.data.slug === `ganti-${run}`, r.data);
r = await call('GET', `/api/news/${art.ID}`);
check('artikel ikut nama kategori baru', r.data.category === `Ganti ${run}` && r.data.categorySlug === `ganti-${run}`, [r.data.category, r.data.categorySlug]);
r = await call('GET', '/api/categories?with_count=1');
check('jumlah artikel per kategori', r.data.find((c) => c.slug === `ganti-${run}`)?.count === 1, r.data.find((c) => c.slug === `ganti-${run}`));

// ---------- komentar ----------
check('komentar tanpa login ditolak', (await call('POST', '/api/comments', { body: { news_id: art.ID, content: 'halo' } })).status === 401);
const cm = await call('POST', '/api/comments', { cookie: user, body: { news_id: art.ID, content: 'Artikel yang bagus sekali <b>x</b>' } });
check('kirim komentar', cm.status === 201, cm.data);
const reply = await call('POST', '/api/comments', { cookie: admin, body: { news_id: art.ID, content: 'Terima kasih banyak', parent_id: cm.data.data?.ID } });
check('balas komentar', reply.status === 201, reply.data);
r = await call('GET', `/api/comments/news/${art.ID}`);
check('komentar tampil beserta balasan', r.status === 200 && r.data.length === 1 && r.data[0].replies?.length === 1, r.data);
check('suka komentar', (await call('POST', `/api/comments/${cm.data.data?.ID}/like`, { cookie: admin })).data.liked === true);
check('batal suka', (await call('POST', `/api/comments/${cm.data.data?.ID}/like`, { cookie: admin })).data.liked === false);
r = await call('GET', '/api/comments', { cookie: admin });
check('daftar komentar admin memuat judul berita', r.data.find((c) => c.ID === cm.data.data?.ID)?.news_title === `${title} Baru`, r.data?.[0]);

r = await call('GET', '/api/admin/stats', { cookie: admin });
check('statistik dashboard', r.status === 200 && r.data.news.total === r.data.news.published + r.data.news.draft && r.data.news.draft >= 1 && r.data.comments.total >= 2, r.data);

// ---------- akun pengguna ----------
check('simpan berita', (await call('POST', '/api/user/bookmarks', { cookie: user, body: { news_id: art.ID } })).data.bookmarked === true);
check('status simpanan', (await call('GET', `/api/user/bookmarks/${art.ID}`, { cookie: user })).data.bookmarked === true);
r = await call('GET', '/api/user/bookmarks', { cookie: user });
check('daftar simpanan memuat judul', r.data[0]?.news_title === `${title} Baru` && r.data[0]?.news_slug === `slug-baru-${run}`, r.data);
check('hapus simpanan', (await call('POST', '/api/user/bookmarks', { cookie: user, body: { news_id: art.ID } })).data.bookmarked === false);
check('catat riwayat', (await call('POST', '/api/user/history', { cookie: user, body: { news_id: art.ID } })).status === 200);
r = await call('GET', '/api/user/history', { cookie: user });
check('daftar riwayat', r.data.length === 1 && r.data[0].news_title === `${title} Baru`, r.data);
check('riwayat tidak bocor ke pengguna lain', !((await call('GET', '/api/user/history', { cookie: admin })).data || []).some((h) => h.ID === r.data[0]?.ID));
check('hapus riwayat', (await call('DELETE', '/api/user/history', { cookie: user })).status === 200 && (await call('GET', '/api/user/history', { cookie: user })).data.length === 0);
r = await call('GET', '/api/user/notifications', { cookie: user });
check('notifikasi balasan masuk', r.data.some((x) => x.type === 'reply') && r.data.some((x) => x.type === 'system'), r.data.map?.((x) => x.type));
check('tandai semua dibaca', (await call('POST', '/api/user/notifications/read-all', { cookie: user })).status === 200 && (await call('GET', '/api/user/notifications', { cookie: user })).data.every((x) => x.is_read));
check('bersihkan notifikasi', (await call('DELETE', '/api/user/notifications', { cookie: user })).status === 200 && (await call('GET', '/api/user/notifications', { cookie: user })).data.length === 0);
check('ubah profil', (await call('PUT', '/api/auth/profile', { cookie: user, body: { name: 'Nama Baru', bio: 'Bio uji', avatar: '' } })).status === 200);
r = await call('GET', '/api/auth/profile', { cookie: user });
check('profil tersimpan', r.data.name === 'Nama Baru' && r.data.bio === 'Bio uji', r.data);
check('ganti password', (await call('PUT', '/api/auth/password', { cookie: user, body: { oldPassword: 'rahasia123', newPassword: 'rahasia456' } })).status === 200);
check('password lama tak berlaku', (await call('POST', '/api/auth/login', { body: { email, password: 'rahasia123' } })).status === 401);
check('password baru berlaku', (await call('POST', '/api/auth/login', { body: { email, password: 'rahasia456' } })).status === 200);

// ---------- poster ----------
const bn = await call('POST', '/api/admin/banners', { cookie: admin, body: { title: 'Poster Uji', image_url: up.data.url, target_url: 'https://contoh.id', is_active: true } });
check('buat poster', bn.status === 201 && bn.data.id > 0, bn.data);
check('poster tampil publik', (await call('GET', '/api/banners')).data.some((b) => b.id === bn.data.id));
check('nonaktifkan poster', (await call('PUT', `/api/admin/banners/${bn.data.id}/toggle`, { cookie: admin })).data.is_active === false && !(await call('GET', '/api/banners')).data.some((b) => b.id === bn.data.id));
const off = await call('POST', '/api/admin/banners', { cookie: admin, body: { title: 'Poster Nonaktif', image_url: up.data.url, is_active: false } });
check('poster dibuat nonaktif tetap nonaktif', off.data.is_active === false && !(await call('GET', '/api/banners')).data.some((b) => b.id === off.data.id), off.data);
check('hapus poster', (await call('DELETE', `/api/admin/banners/${bn.data.id}`, { cookie: admin })).status === 200);

// ---------- event & tiket ----------
const evt = await call('POST', '/api/admin/events', { cookie: admin, body: { title: `Event Uji ${run}`, slug: `event-uji-${run}`, description: 'Deskripsi', venue: 'Zoom', start_time: new Date(Date.now() + 864e5).toISOString(), end_time: new Date(Date.now() + 9e7).toISOString(), quota: 1, price: 0, status: 'published', speakers: [{ name: 'Pemateri', role: 'Dosen' }] } });
check('buat event', evt.status < 300, evt.data);
r = await call('GET', '/api/events');
const ev = (r.data.data || []).find((e) => e.slug === `event-uji-${run}`);
check('daftar event', !!ev && ev.speakers?.length === 1, r.data);
check('detail event', (await call('GET', `/api/events/slug/event-uji-${run}`)).data?.data?.ID === ev?.ID);
const tick = await call('POST', `/api/events/${ev?.ID}/register`, { cookie: user, body: { name: 'Pengguna Uji', email, phone: '08', tier: 'free', total_amount: 0 } });
check('daftar event', tick.status === 200 && /^TIC-/.test(tick.data.ticket_code || ''), tick.data);
check('kuota penuh ditolak', (await call('POST', `/api/events/${ev?.ID}/register`, { body: { name: 'Lain', email: 'lain@contoh.id', tier: 'free' } })).status === 400);
r = await call('GET', '/api/user/tickets', { cookie: user });
check('tiket pengguna memuat event', r.data.data?.[0]?.event?.slug === `event-uji-${run}`, r.data);
check('pendaftaran terlihat admin', (await call('GET', '/api/admin/registrations', { cookie: admin })).data.data?.length >= 1);
check('hapus event', (await call('DELETE', `/api/admin/events/${ev?.ID}`, { cookie: admin })).status === 200);

// ---------- CORS ----------
r = await call('OPTIONS', '/api/news', { headers: { Origin: 'http://localhost:4321', 'Access-Control-Request-Method': 'PUT' } });
check('CORS frontend diizinkan', r.headers.get('access-control-allow-origin') === 'http://localhost:4321' && r.headers.get('access-control-allow-credentials') === 'true');
r = await call('GET', '/api/news?limit=1', { headers: { Origin: 'https://situs-jahat.example' } });
check('CORS situs lain ditolak', r.headers.get('access-control-allow-origin') !== 'https://situs-jahat.example');

// ---------- bersih-bersih ----------
check('hapus komentar', (await call('DELETE', `/api/comments/${cm.data.data?.ID}`, { cookie: admin })).status === 200);
for (const id of [art.ID, twin.data.data?.ID, draftArt.ID]) check(`hapus berita #${id}`, (await call('DELETE', `/api/news/${id}`, { cookie: admin })).status === 200);
check('berita terhapus hilang', (await call('GET', `/api/news/slug/slug-baru-${run}`)).status === 404);
check('berita terhapus tak bisa dialihkan', (await call('GET', `/api/news/resolve/${art.slug}`)).status === 404);
for (const c of [cat, cat2]) check(`hapus kategori #${c.data.ID}`, (await call('DELETE', `/api/categories/${c.data.ID}`, { cookie: admin })).status === 200);
check('logout menghapus cookie', /auth_token=;/.test((await call('POST', '/api/auth/logout')).headers.get('set-cookie') || ''));

console.log(`\nLULUS ${pass}, GAGAL ${failures.length}`);
failures.forEach((f) => console.log('  ✗ ' + f));
process.exit(failures.length ? 1 : 0);

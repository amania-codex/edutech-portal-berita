// Memeriksa bahwa semua berita terbit tersedia dan semua gambarnya benar-benar bisa dimuat.
//   node tools/check-images.mjs        (BASE_URL=http://localhost:3000, MIN_NEWS=0)
const BASE = process.env.BASE_URL || 'http://localhost:3000';
const MIN = Number(process.env.MIN_NEWS || 0);
// Memang tidak ada di arsip WordPress lama
const KNOWN_MISSING = new Set(['/uploads/wp/2025/05/suyanto.pptx', '/uploads/wp/2025/08/sekda-bms1-300x228.webp', '/uploads/wp/2025/09/mui-300x169.webp']);

const urls = new Map();
let news = 0;
for (let offset = 0; ; offset += 100) {
  const list = await (await fetch(`${BASE}/api/news?limit=100&offset=${offset}`)).json();
  for (const n of list) {
    news++;
    const found = [...n.content.matchAll(/(?:src|href)="(\/uploads\/[^"]+)"/g)].map((m) => m[1].replaceAll('&amp;', '&'));
    if (n.imageUrl.startsWith('/uploads/')) found.push(encodeURI(n.imageUrl));
    for (const u of found) if (!urls.has(u)) urls.set(u, n.slug);
  }
  if (list.length < 100) break;
}

const bad = [];
const queue = [...urls];
await Promise.all(Array.from({ length: 24 }, async () => {
  for (let item; (item = queue.pop());) {
    const [url, slug] = item;
    if (KNOWN_MISSING.has(decodeURI(url))) continue;
    try {
      const res = await fetch(BASE + url);
      const size = (await res.arrayBuffer()).byteLength;
      if (res.status !== 200 || size === 0) bad.push(`${res.status} ${url} (${slug})`);
    } catch (e) {
      bad.push(`${e.message} ${url} (${slug})`);
    }
  }
}));

console.log(`Berita terbit: ${news}, file diperiksa: ${urls.size}, gagal: ${bad.length}`);
bad.slice(0, 20).forEach((b) => console.log('  ✗ ' + b));
if (process.env.GITHUB_ACTIONS) console.log(`::${bad.length || news < MIN ? 'error' : 'notice'}::Berita ${news}, file ${urls.size}, gagal ${bad.length}%0A${bad.slice(0, 10).join('%0A')}`);
if (news < MIN) console.log(`  ✗ jumlah berita kurang dari ${MIN}`);
process.exit(bad.length || news < MIN ? 1 : 0);

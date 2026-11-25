/**
 * Merapikan isi berita sebelum ditampilkan (terutama arsip WordPress):
 * - subjudul pembuka (<p><strong>*...</strong></p>) dipindah ke bawah judul
 * - foto + keterangan miring digabung menjadi <figure> dengan <figcaption>
 * - foto pertama yang sama dengan foto utama dibuang; keterangannya dipakai untuk foto utama
 * - baris tebal di awal paragraf (<strong>Judul</strong><br>) menjadi subjudul <h3>
 * - paragraf kosong dibuang
 * Fungsi murni (tanpa impor) agar bisa diuji langsung dengan Node.
 */

export interface PreparedArticle {
  html: string;
  subtitle: string;
  heroCaption: string;
}

export interface ReadAlsoLink {
  title: string;
  href: string;
}

const CONTAINER = /<(\/?)(table|blockquote|ul|ol|div|figure|pre|center|iframe|script)\b[^>]*>/gi;
const INLINE_WRAPPER = /<\/?(?:a|strong|b|span|em|i|u)\b[^>]*>/gi;

/** Nama berkas tanpa ukuran WordPress (-1024x511), -scaled, dan ekstensi. */
export function imageStem(url: string): string {
  let name = String(url || '').split(/[?#]/)[0].split('/').pop() || '';
  try {
    name = decodeURIComponent(name);
  } catch {}
  return name.replace(/\.\w+$/, '').replace(/(-scaled)?(-\d+x\d+)?$/, '').toLowerCase();
}

const textOf = (html: string) =>
  html
    .replace(/<[^>]+>/g, '')
    .replace(/&nbsp;|&#160;| /g, ' ')
    .replace(/\s+/g, ' ')
    .trim();

/** Bagian yang bersarang (tabel, kutipan, daftar, embed) disimpan utuh agar tidak ikut diubah. */
function protect(html: string): { text: string; saved: string[] } {
  const saved: string[] = [];
  let text = '';
  let depth = 0;
  let start = 0;
  let last = 0;
  let open = '';
  CONTAINER.lastIndex = 0;
  for (let m: RegExpExecArray | null; (m = CONTAINER.exec(html)); ) {
    const tag = m[2].toLowerCase();
    if (!m[1]) {
      if (depth === 0) {
        text += html.slice(last, m.index);
        start = m.index;
        open = tag;
      }
      if (tag === open) depth++;
    } else if (depth > 0 && tag === open && --depth === 0) {
      const end = m.index + m[0].length;
      saved.push(html.slice(start, end));
      text += `\u0000${saved.length - 1}\u0000`;
      last = end;
    }
  }
  text += html.slice(depth > 0 ? start : last);
  return { text, saved };
}

const restore = (html: string, saved: string[]) => html.replace(/\u0000(\d+)\u0000/g, (_m, i) => saved[Number(i)]);

interface Block {
  kind: 'p' | 'other';
  html: string;
  inner: string;
}

function toBlocks(text: string): Block[] {
  return text
    .split(/(<p\b[^>]*>[\s\S]*?<\/p>)/i)
    .filter((s) => s.trim() !== '')
    .map((html) => {
      const m = html.match(/^<p\b[^>]*>([\s\S]*?)<\/p>$/i);
      return m ? { kind: 'p', html, inner: m[1] } : { kind: 'other', html, inner: '' };
    });
}

/** Buang pembungkus inline yang jadi kosong dan <br> di awal/akhir. */
function trimInline(html: string): string {
  let out = html.replace(/&nbsp;/g, ' ');
  for (let prev = ''; prev !== out; ) {
    prev = out;
    out = out
      .replace(/<(a|strong|b|span|em|i|u)\b[^>]*>\s*(?:<br\s*\/?>\s*)*<\/\1>/gi, '')
      .replace(/^\s*(?:<br\s*\/?>\s*)+|(?:\s*<br\s*\/?>)+\s*$/gi, '')
      .trim();
  }
  return out;
}

/**
 * Paragraf yang diawali satu foto. Teks di belakangnya: keterangan miring → caption,
 * selain itu tetap jadi paragraf (rest).
 */
function imageParagraph(inner: string): { img: string; caption: string; rest: string } | null {
  const imgs = inner.match(/<img\b[^>]*>/gi);
  if (!imgs || imgs.length !== 1) return null;
  const at = inner.indexOf(imgs[0]);
  if (textOf(inner.slice(0, at)) !== '') return null;
  // Tautan yang hanya membungkus foto (ke berkas aslinya) ikut dibuang
  const rest = trimInline(inner.replace(/<a\b[^>]*>\s*(<img\b[^>]*>)\s*<\/a>/i, '$1').replace(imgs[0], ''));
  if (textOf(rest) === '') return { img: imgs[0], caption: '', rest: '' };
  const cap = rest.replace(/<\/?(?:strong|b|span)\b[^>]*>/gi, '').trim().match(/^<(em|i)\b[^>]*>([\s\S]*?)<\/\1>$/i);
  if (cap && !/<\/?(em|i)\b/i.test(cap[2]) && textOf(cap[2]).length <= 400) {
    return { img: imgs[0], caption: cap[2].replace(/<[^>]+>/g, '').trim(), rest: '' };
  }
  return { img: imgs[0], caption: '', rest };
}

/** Paragraf yang seluruhnya teks miring pendek: keterangan foto di atasnya. */
function captionParagraph(inner: string): string | null {
  const m = inner.replace(/&nbsp;/g, ' ').trim().match(/^<(em|i)\b[^>]*>([\s\S]*?)<\/\1>$/i);
  if (!m || /<img\b/i.test(m[2])) return null;
  const text = textOf(m[2]);
  return text && text.length <= 400 ? m[2].replace(/<[^>]+>/g, '').trim() : null;
}

/** Teks tebal pendek yang layak jadi subjudul (bukan kalimat penutup, bukan inisial penulis). */
function headingText(text: string): boolean {
  return text.length >= 3 && text.length <= 120 && !/^\(.*\)$/.test(text) && !/[.,;"”]$/.test(text) && !isByline(text);
}

/** Baris penulis/dateline: "Oleh: ...", nama bergelar (S.Pd, M.Si), "KOTA, EDUKATOR". */
function isByline(text: string): boolean {
  return /^\s*oleh\b/i.test(text) || /\b[SM]\.\s?[A-Z][A-Za-z]*\.?/.test(text) || /edukator\s*[–-]?\s*$/i.test(text);
}

const figure = (img: string, caption: string) =>
  `<figure class="content-figure">${img}${caption ? `<figcaption>${caption}</figcaption>` : ''}</figure>`;

export function prepareArticle(content: string, heroUrl: string): PreparedArticle {
  const { text, saved } = protect(String(content || ''));
  const blocks = toBlocks(text);
  let subtitle = '';
  let heroCaption = '';
  const out: string[] = [];
  const heroStem = imageStem(heroUrl);
  let seenImage = false;
  let first = true;

  for (let i = 0; i < blocks.length; i++) {
    const b = blocks[i];
    if (b.kind === 'other') {
      const full = restore(b.html, saved);
      // Foto ber-keterangan bawaan WordPress (<figure>) yang sama dengan foto utama
      const fig = full.trim().match(/^<figure\b[^>]*>\s*(?:<a\b[^>]*>\s*)?(<img\b[^>]*>)(?:\s*<\/a>)?\s*(?:<figcaption\b[^>]*>([\s\S]*?)<\/figcaption>)?\s*<\/figure>$/i);
      if (fig && !seenImage && heroStem && imageStem(fig[1].match(/\bsrc="([^"]+)"/i)?.[1] || '') === heroStem) {
        heroCaption = textOf(fig[2] || '');
        seenImage = true;
        first = false;
        continue;
      }
      out.push(b.html);
      if (/<img\b/i.test(full)) seenImage = true;
      if (textOf(restore(b.html, saved)) !== '' || /\u0000/.test(b.html)) first = false;
      continue;
    }
    const plain = textOf(b.inner);
    if (plain === '' && !/<(img|iframe|video|audio|embed)\b/i.test(b.inner) && !/\u0000/.test(b.inner)) continue;

    // Subjudul pembuka
    const kicker = b.inner.trim().match(/^<(strong|b)\b[^>]*>([^<]{3,250})<\/\1>$/i);
    if (first && kicker && !subtitle && !isByline(textOf(kicker[2]))) {
      subtitle = textOf(kicker[2]).replace(/^\*+\s*/, '');
      continue;
    }
    // Subjudul pembuka yang menempel dengan foto: <strong>*Subjudul<img ...></strong>
    const glued = first && !subtitle && b.inner.match(/^\s*<(strong|b)\b[^>]*>([^<]{3,250})(<\/\1>)?\s*(?:<br\s*\/?>\s*)*(?=<img\b|<a\b[^>]*>\s*<img\b)/i);
    if (glued && !isByline(textOf(glued[2]))) {
      subtitle = textOf(glued[2]).replace(/^\*+\s*/, '');
      let restInner = b.inner.slice(glued[0].length);
      if (!glued[3]) restInner = restInner.replace(new RegExp(`</${glued[1]}>`, 'i'), '');
      blocks[i] = { kind: 'p', html: `<p>${restInner}</p>`, inner: restInner };
      i--;
      first = false;
      continue;
    }
    first = false;

    // Foto (+ keterangan)
    const pic = imageParagraph(b.inner);
    if (pic) {
      let caption = pic.caption;
      const next = blocks[i + 1];
      if (!caption && !pic.rest && next?.kind === 'p') {
        const c = captionParagraph(next.inner);
        if (c) {
          caption = c;
          i++;
        }
      }
      const src = pic.img.match(/\bsrc="([^"]+)"/i)?.[1] || '';
      const rest = pic.rest ? `<p>${pic.rest}</p>` : '';
      if (!seenImage && heroStem && imageStem(src) === heroStem) {
        heroCaption = caption;
        seenImage = true;
        if (rest) blocks.splice(i + 1, 0, { kind: 'p', html: rest, inner: pic.rest });
        continue;
      }
      seenImage = true;
      out.push(figure(pic.img, caption));
      // Sisa teks diproses lagi (bisa berupa subjudul)
      if (rest) blocks.splice(i + 1, 0, { kind: 'p', html: rest, inner: pic.rest });
      continue;
    }
    if (/<img\b/i.test(b.inner)) seenImage = true;

    // Subjudul di dalam tulisan
    const strongOnly = b.inner.trim().match(/^<(strong|b)\b[^>]*>([^<]+)<\/\1>$/i);
    if (strongOnly && headingText(textOf(strongOnly[2]))) {
      out.push(`<h3>${strongOnly[2].trim()}</h3>`);
      continue;
    }
    const lead = b.inner.match(/^\s*<(strong|b)\b[^>]*>([^<]+)<\/\1>\s*<br\s*\/?>\s*([\s\S]+)$/i);
    if (lead && headingText(textOf(lead[2])) && textOf(lead[3]) !== '') {
      out.push(`<h3>${lead[2].trim()}</h3>`, `<p>${lead[3].trim()}</p>`);
      continue;
    }
    out.push(b.html);
  }

  return { html: restore(out.join('\n'), saved), subtitle, heroCaption };
}

const escapeHtml = (s: string) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

/**
 * Kotak "Baca Juga" disisipkan setelah paragraf ke-`after` (paragraf berisi teks),
 * seperti portal berita besar. Tidak disisipkan bila tulisan sudah punya "Baca juga" sendiri
 * atau terlalu pendek.
 */
export function insertReadAlso(html: string, links: ReadAlsoLink[], after = 3): string {
  if (!links.length || /baca\s+juga/i.test(html)) return html;
  const box =
    `<aside class="read-also"><span class="read-also-label">Baca Juga</span><ul>` +
    links.map((l) => `<li><a href="${escapeHtml(l.href)}">${escapeHtml(l.title)}</a></li>`).join('') +
    `</ul></aside>`;
  let count = 0;
  let done = false;
  // Hanya paragraf di tingkat atas (bukan di dalam tabel/kutipan)
  const { text, saved } = protect(html);
  const result = text.replace(/<p\b[^>]*>[\s\S]*?<\/p>/gi, (p) => {
    if (done || textOf(p).length < 60) return p;
    if (++count === after) {
      done = true;
      return p + box;
    }
    return p;
  });
  return done ? restore(result, saved) : html;
}

/**
 * Data tampilan situs yang diatur dari dashboard: poster, halaman, pengaturan, dan menu.
 * Dipakai di server saat halaman dibuat; disimpan 30 detik agar tidak memanggil API berulang.
 */
import { apiGet, assetUrl, cached, getCategories, type Category } from './news';

export interface Banner {
  id: number;
  title: string;
  imageUrl: string;
  targetUrl: string;
  placement: 'header' | 'tengah' | 'kiri' | 'kanan';
}

export interface SitePage {
  title: string;
  slug: string;
  menuOrder: number;
}

export type Settings = Record<string, string>;

export const DEFAULT_DESCRIPTION =
  'Portal informasi pendidikan, teknologi, riset, dan inovasi. Menghubungkan gagasan dengan perubahan melalui jurnalisme bermutu dan terpercaya.';

// Gagal memuat bagian ini tidak boleh membuat seluruh halaman error
const safe = <T>(load: () => Promise<T>, fallback: T) => () => load().catch(() => fallback);

export const getSettings = cached(5, safe<Settings>(() => apiGet('/settings'), {}));

export const getPages = cached(
  5,
  safe<SitePage[]>(async () => (await apiGet('/pages')).map((p: any) => ({ title: p.title, slug: p.slug, menuOrder: p.menuOrder ?? 0 })), []),
);

export const getBanners = cached(
  5,
  safe<Banner[]>(
    async () =>
      (await apiGet('/banners')).map((b: any) => ({
        id: b.id,
        title: b.title,
        imageUrl: assetUrl(b.image_url),
        targetUrl: b.target_url || '',
        placement: b.placement || 'kanan',
      })),
    [],
  ),
);

export async function bannersAt(placement: Banner['placement']): Promise<Banner[]> {
  return (await getBanners()).filter((b) => b.placement === placement);
}

/** Kategori di menu navigasi, sesuai urutan yang diatur admin. */
export async function getMenuCategories(): Promise<Category[]> {
  const categories = await getCategories().catch(() => [] as Category[]);
  return categories.filter((c) => c.menuOrder > 0).sort((a, b) => a.menuOrder - b.menuOrder || a.name.localeCompare(b.name));
}

/** Kategori bertanda "Topik Pilihan" untuk beranda. */
export async function getTopicCategories(): Promise<Array<Category & { parentName: string }>> {
  const categories = await getCategories().catch(() => [] as Category[]);
  return categories
    .filter((c) => c.isTopic)
    .sort((a, b) => (a.menuOrder || 99) - (b.menuOrder || 99) || a.name.localeCompare(b.name))
    .map((c) => ({ ...c, parentName: categories.find((p) => p.id === c.parentId)?.name || '' }));
}

export const SOCIAL_LINKS: Array<{ key: string; label: string; icon: string }> = [
  { key: 'facebook_url', label: 'Facebook', icon: 'f' },
  { key: 'x_url', label: 'X / Twitter', icon: '𝕏' },
  { key: 'instagram_url', label: 'Instagram', icon: '◎' },
  { key: 'youtube_url', label: 'YouTube', icon: '▶' },
  { key: 'tiktok_url', label: 'TikTok', icon: '♪' },
];

/** Tautan poster: alamat luar dibuka di tab baru. */
export function bannerLink(b: Banner) {
  const external = /^https?:\/\//i.test(b.targetUrl);
  return { href: b.targetUrl || undefined, target: external ? '_blank' : undefined, rel: external ? 'noopener sponsored' : undefined };
}

const escapeHtml = (s: string) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

/**
 * Isi halaman informasi (teks biasa) → HTML aman: paragraf, subjudul "## ",
 * baris baru, serta tautan web/email otomatis. HTML yang diketik admin tidak dijalankan.
 */
export function textToHtml(text: string): string {
  const linkify = (s: string) =>
    s
      .replace(/https?:\/\/[^\s<]+[^\s<.,;:!?)\]'"]/g, (u) => `<a href="${u}" target="_blank" rel="noopener">${u}</a>`)
      .replace(/(^|[\s(])([\w.+-]+@[\w-]+(?:\.[\w-]+)+)/g, (_m, pre, mail) => `${pre}<a href="mailto:${mail}">${mail}</a>`);
  return text
    .replace(/\r\n?/g, '\n')
    .split(/\n\s*\n/)
    .map((block) => block.trim())
    .filter(Boolean)
    .map((block) => {
      const para = (t: string) => (t.trim() ? `<p>${linkify(escapeHtml(t.trim())).replace(/\n/g, '<br />')}</p>` : '');
      if (!block.startsWith('## ')) return para(block);
      const [first, ...rest] = block.split('\n');
      return `<h2>${linkify(escapeHtml(first.slice(3).trim()))}</h2>${para(rest.join('\n'))}`;
    })
    .join('\n');
}

export interface Video {
  id: number;
  title: string;
  youtubeId: string;
  description: string;
}

/** Video YouTube dari Dashboard > Kelola Video (sudah urut sesuai pengaturan). */
export const getVideos = cached(
  5,
  safe<Video[]>(
    async () =>
      (await apiGet('/videos?limit=48')).map((v: any) => ({
        id: v.ID ?? v.id,
        title: v.title || '',
        youtubeId: v.youtubeId || '',
        description: v.description || '',
      })),
    [],
  ),
);

export const videoThumb = (id: string, size: 'hq' | 'mq' | 'maxres' = 'hq') =>
  `https://i.ytimg.com/vi/${id}/${size === 'maxres' ? 'maxresdefault' : `${size}default`}.jpg`;

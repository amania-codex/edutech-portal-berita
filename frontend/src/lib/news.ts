/**
 * Pengambilan & normalisasi data berita untuk halaman publik.
 * Dipakai saat build (Astro static) maupun di browser (pencarian).
 */
import { API_BASE, BACKEND_URL } from './config';

export const FALLBACK_IMAGE =
  'https://images.unsplash.com/photo-1504711434969-e33886168f5c?auto=format&fit=crop&w=1200&q=80';

export interface News {
  id: number;
  title: string;
  slug: string;
  excerpt: string;
  content: string;
  imageUrl: string;
  category: string;
  categorySlug: string;
  categories: string[];
  author: string;
  tags: string[];
  viewCount: number;
  readTime: number;
  isFeatured: boolean;
  createdAt: string;
}

export interface Category {
  id: number;
  name: string;
  slug: string;
  parentId: number | null;
}

/** Path /uploads/... disajikan oleh backend, bukan oleh host frontend. */
export function assetUrl(path?: string): string {
  if (!path) return '';
  return path.startsWith('/uploads/') ? BACKEND_URL + encodeURI(path) : path;
}

export function slugify(s: string): string {
  return s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '');
}

export function normalizeNews(a: any): News {
  const category = a.category || '';
  const categorySlug = a.categorySlug || slugify(category);
  return {
    id: a.ID ?? a.id ?? 0,
    title: a.title || '',
    slug: a.slug || '',
    excerpt: a.excerpt || '',
    content: a.content || '',
    imageUrl: assetUrl(a.imageUrl || a.image_url) || FALLBACK_IMAGE,
    category,
    categorySlug,
    categories: [...new Set<string>([...(a.categories || []), categorySlug])].filter(Boolean),
    author: a.author || 'Redaksi',
    tags: String(a.tags || '').split(',').map((t) => t.trim()).filter(Boolean),
    viewCount: a.viewCount ?? 0,
    readTime: a.readTime || 1,
    isFeatured: !!a.isFeatured,
    createdAt: a.CreatedAt || a.createdAt || '',
  };
}

const imageStem = (url: string) => decodeURI(url).replace(/(-scaled)?(-\d+x\d+)?\.\w+$/, '');

/** Konten siap tampil: URL gambar absolut, lazy-load, tanpa foto pembuka yang sama dengan foto utama. */
export function renderContent(html: string, heroUrl: string): string {
  let out = html.replace(/\b(src|href)="(\/uploads\/[^"]*)"/g, (_m, attr, path) => `${attr}="${BACKEND_URL}${path}"`);
  const lead = out.match(/^\s*(?:<(?:p|a|div|center|span|strong|em|b|i|h\d)\b[^>]*>\s*)*(<img\b[^>]*\bsrc="([^"]+)"[^>]*>)/);
  if (lead && imageStem(lead[2]) === imageStem(heroUrl)) {
    out = out
      .replace(lead[1], '')
      // buang pembungkus yang jadi kosong, dari dalam ke luar
      .replace(/^(\s*<(?:p|div|center|h\d)\b[^>]*>\s*)<(a|span|strong|em|b|i)\b[^>]*>\s*<\/\2>/, '$1')
      .replace(/^\s*<(p|div|center|h\d)\b[^>]*>\s*<\/\1>/, '');
  }
  return out.replace(/<img\b(?![^>]*\bloading=)/g, '<img loading="lazy"');
}

/** GET ke API. Gagal = error, supaya build tidak diam-diam menghasilkan situs kosong. */
export async function apiGet(path: string): Promise<any> {
  const res = await fetch(`${API_BASE}${path}`);
  if (!res.ok) throw new Error(`API ${path} gagal (${res.status})`);
  return res.json();
}

export async function getNews(path: string): Promise<News[]> {
  const data = await apiGet(path);
  return (Array.isArray(data) ? data : []).map(normalizeNews);
}

function cached<T>(load: () => Promise<T>): () => Promise<T> {
  let at = 0;
  let value: Promise<T> | null = null;
  return () => {
    // saat build cukup sekali; di dev server disegarkan tiap 30 detik
    if (!value || (import.meta.env.DEV && Date.now() - at > 30_000)) {
      at = Date.now();
      value = load().catch((e) => {
        value = null;
        throw e;
      });
    }
    return value;
  };
}

/** Semua artikel terbit, terbaru lebih dulu (API membatasi 100 per permintaan). */
export const getAllNews = cached(async (): Promise<News[]> => {
  const all = new Map<number, News>();
  for (let offset = 0; ; offset += 100) {
    const batch = await getNews(`/news?limit=100&offset=${offset}`);
    batch.forEach((n) => all.set(n.id, n));
    if (batch.length < 100) return [...all.values()];
  }
});

export const getCategories = cached(async (): Promise<Category[]> =>
  (await apiGet('/categories')).map((c: any) => ({
    id: c.ID ?? c.id,
    name: c.name,
    slug: c.slug,
    parentId: c.parentId ?? null,
  })),
);

export const getTrending = cached(() => getNews('/news/trending?limit=5'));

/** Slug kategori beserta seluruh sub-kategorinya. */
export function withDescendants(slug: string, categories: Category[]): string[] {
  const root = categories.find((c) => c.slug === slug);
  if (!root) return [slug];
  const found = [root];
  for (let i = 0; i < found.length; i++) {
    found.push(...categories.filter((c) => c.parentId === found[i].id && !found.includes(c)));
  }
  return found.map((c) => c.slug);
}

export function newsInCategory(slug: string, news: News[], categories: Category[]): News[] {
  if (slug === 'semua') return news;
  const slugs = new Set(withDescendants(slug, categories));
  return news.filter((n) => n.categories.some((c) => slugs.has(c)));
}

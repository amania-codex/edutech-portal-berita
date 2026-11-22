/**
 * Helper untuk script di browser (dashboard & halaman akun).
 * Situs ini statis: data milik pengguna diambil setelah halaman dibuka, memakai cookie login.
 */
import { API_BASE } from './config';

export const el = (id: string) => document.getElementById(id) as any;

export function api(path: string, init: RequestInit = {}): Promise<Response> {
  return fetch(`${API_BASE}${path}`, { credentials: 'include', ...init });
}

export async function getJSON<T = any>(path: string): Promise<T> {
  const res = await api(path);
  if (!res.ok) throw new Error(`${path}: ${res.status}`);
  return res.json();
}

export function sendJSON(path: string, method: string, body: unknown): Promise<Response> {
  return api(path, { method, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
}

/** Pesan error dari respons API, apa pun nama field-nya. */
export async function errorText(res: Response, fallback: string): Promise<string> {
  const data = await res.json().catch(() => ({}));
  return data.error || data.message || fallback;
}

export function clone(templateId: string): DocumentFragment {
  return (el(templateId) as HTMLTemplateElement).content.cloneNode(true) as DocumentFragment;
}

/** Isi teks elemen; diam bila elemennya tidak ada. */
export function setText(root: ParentNode, selector: string, text: unknown) {
  const target = root.querySelector(selector);
  if (target) target.textContent = String(text ?? '');
}

export function show(target: HTMLElement | null, visible: boolean) {
  if (target) target.style.display = visible ? '' : 'none';
}

export function formatDate(iso?: string, opts: Intl.DateTimeFormatOptions = { day: 'numeric', month: 'short', year: 'numeric' }): string {
  const d = iso ? new Date(iso) : null;
  return d && !isNaN(d.getTime()) ? d.toLocaleDateString('id-ID', { timeZone: 'Asia/Jakarta', ...opts }) : '-';
}

export function relativeTime(iso?: string): string {
  if (!iso) return 'Baru saja';
  const sec = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (sec < 60) return 'Baru saja';
  if (sec < 3600) return `${Math.floor(sec / 60)} menit lalu`;
  if (sec < 86400) return `${Math.floor(sec / 3600)} jam lalu`;
  if (sec < 172800) return 'Kemarin';
  return formatDate(iso, { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' });
}

export const createdAt = (row: any): string => row.CreatedAt || row.created_at || row.createdAt || '';
export const rowId = (row: any): number => row.ID ?? row.id;

let toastTimer: ReturnType<typeof setTimeout>;
export function toast(id: string, message: string) {
  const box = el(id) as HTMLElement | null;
  if (!box) return;
  box.textContent = message;
  box.style.display = 'block';
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { box.style.display = 'none'; }, 3000);
}

/** Hanya tautan http(s)/relatif yang boleh dipasang ke href (cegah javascript:). */
export function safeHref(url?: string): string {
  return url && /^(https?:\/\/|\/(?!\/))/i.test(url) ? url : '#';
}

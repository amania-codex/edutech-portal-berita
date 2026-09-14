/**
 * NEWS CONTROLLER
 * Handles news CRUD business logic.
 * Architecture: MVC - Controller layer.
 */
import { NewsModel, type CreateNewsData } from '../models/NewsModel';

export interface NewsResult {
  success: boolean;
  error?: string;
  data?: any;
  id?: number;
}

export class NewsController {
  static getAll(options: { limit?: number; offset?: number; status?: string; category?: string } = {}): NewsResult {
    try {
      const articles = NewsModel.getAll(options);
      return { success: true, data: articles };
    } catch (err) {
      return { success: false, error: 'Gagal mengambil data berita.' };
    }
  }

  static getFeatured(limit = 5): NewsResult {
    try {
      const articles = NewsModel.getFeatured(limit);
      return { success: true, data: articles };
    } catch (err) {
      return { success: false, error: 'Gagal mengambil berita unggulan.' };
    }
  }

  static getTrending(limit = 5): NewsResult {
    try {
      const articles = NewsModel.getTrending(limit);
      return { success: true, data: articles };
    } catch (err) {
      return { success: false, error: 'Gagal mengambil berita trending.' };
    }
  }

  static getBySlug(slug: string): NewsResult {
    try {
      const article = NewsModel.findBySlug(slug);
      if (!article) return { success: false, error: 'Berita tidak ditemukan.' };
      NewsModel.incrementViews(article.id);
      return { success: true, data: article };
    } catch (err) {
      return { success: false, error: 'Gagal mengambil berita.' };
    }
  }

  static getById(id: number): NewsResult {
    try {
      const article = NewsModel.findById(id);
      if (!article) return { success: false, error: 'Berita tidak ditemukan.' };
      return { success: true, data: article };
    } catch (err) {
      return { success: false, error: 'Gagal mengambil berita.' };
    }
  }

  static search(query: string): NewsResult {
    try {
      const articles = NewsModel.search(query);
      return { success: true, data: articles };
    } catch (err) {
      return { success: false, error: 'Gagal mencari berita.' };
    }
  }

  static create(data: CreateNewsData): NewsResult {
    // Validation
    if (!data.title?.trim()) return { success: false, error: 'Judul wajib diisi.' };
    if (!data.content?.trim()) return { success: false, error: 'Konten wajib diisi.' };
    if (!data.category?.trim()) return { success: false, error: 'Kategori wajib diisi.' };

    // Generate slug if not provided
    if (!data.slug) {
      data.slug = NewsModel.generateSlug(data.title);
    }

    // Ensure slug is unique
    let slug = data.slug;
    let counter = 1;
    while (NewsModel.findBySlug(slug)) {
      slug = `${data.slug}-${counter++}`;
    }
    data.slug = slug;

    try {
      const id = NewsModel.create(data);
      return { success: true, id };
    } catch (err: any) {
      if (err.message?.includes('UNIQUE constraint')) {
        return { success: false, error: 'Slug sudah digunakan. Coba judul yang berbeda.' };
      }
      return { success: false, error: 'Gagal menyimpan berita.' };
    }
  }

  static update(id: number, data: Partial<CreateNewsData> & { status?: string; is_featured?: number }): NewsResult {
    const existing = NewsModel.findById(id);
    if (!existing) return { success: false, error: 'Berita tidak ditemukan.' };

    if (data.title && !data.slug) {
      data.slug = NewsModel.generateSlug(data.title);
    }

    try {
      NewsModel.update(id, data);
      return { success: true, data: NewsModel.findById(id) };
    } catch (err: any) {
      if (err.message?.includes('UNIQUE constraint')) {
        return { success: false, error: 'Slug sudah digunakan.' };
      }
      return { success: false, error: 'Gagal memperbarui berita.' };
    }
  }

  static delete(id: number): NewsResult {
    const existing = NewsModel.findById(id);
    if (!existing) return { success: false, error: 'Berita tidak ditemukan.' };

    try {
      NewsModel.delete(id);
      return { success: true };
    } catch (err) {
      return { success: false, error: 'Gagal menghapus berita.' };
    }
  }
}
